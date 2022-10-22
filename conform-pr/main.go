// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/github-actions/internal"
	"github.com/FerretDB/github-actions/internal/graphql"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DebugEnv(action)

	ctx := context.Background()
	client := graphql.NewClient(ctx, action, "CONFORM_TOKEN")

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tCheck\tStatus\t\n")
	fmt.Fprintf(w, "\t-----\t------\t\n")

	conform := true

	event, err := internal.ReadEvent(action)
	if err != nil {
		action.Errorf("Failed to read event: %s.", err)
	}

	var results []checkResult

	switch event := event.(type) {
	case *github.PullRequestEvent:
		results = runChecks(ctx, action, client, *event.PullRequest.NodeID)
	default:
		action.Fatalf("Unexpected event type: %T.", event)
	}

	for _, res := range results {
		status := ":white_check_mark:"
		if res.err != nil {
			status = ":x: " + res.err.Error()
			conform = false
		}

		fmt.Fprintf(w, "\t%s\t%s\t\n", res.check, status)
	}

	w.Flush()
	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	if !conform {
		action.Fatalf("The PR does not conform to the rules.")
	}
}

// checkResult is a result of a single check.
type checkResult struct {
	check string
	err   error
}

// runChecks runs all the checks on the given PR GraphQL node.
func runChecks(ctx context.Context, action *githubactions.Action, client *graphql.Client, nodeID string) []checkResult {
	pr := client.GetPullRequest(ctx, nodeID)

	var res []checkResult

	res = append(res, checkResult{
		check: "Labels",
		err:   checkLabels(action, pr.Labels),
	})
	res = append(res, checkResult{
		check: "Size",
		err:   checkSize(action, pr.ProjectFields),
	})
	res = append(res, checkResult{
		check: "Sprint",
		err:   checkSprint(action, pr.ProjectFields),
	})

	// PRs from dependabot have good enough title and body
	if pr.Author == "dependabot" && pr.AuthorBot {
		return res
	}

	res = append(res, checkResult{
		check: "Title",
		err:   checkTitle(action, pr.Title),
	})
	res = append(res, checkResult{
		check: "Body",
		err:   checkBody(action, pr.Body),
	})

	return res
}

// checkLabels checks if PR's labels are valid.
func checkLabels(action *githubactions.Action, labels []string) error {
	if slices.Contains(labels, "no ci") {
		action.Fatalf(`"no ci" label should be handled by configuration, not conform-pr`)
	}

	var res []string

	for _, l := range []string{
		"good first issue",
		"help wanted",
		"not ready",
		"scope changed",

		// temporary labels for issues
		"code/tigris",
		"fuzz",
		"validation",
	} {
		if slices.Contains(labels, l) {
			res = append(res, l)
		}
	}

	if res != nil {
		return fmt.Errorf("Those labels should not be applied to PRs: %s", strings.Join(res, ", "))
	}

	if slices.Contains(labels, "do not merge") {
		return fmt.Errorf("That PR should not be merged yet")
	}

	return nil
}

// checkSize checks that PR does not contain "Size" field with a set value.
func checkSize(_ *githubactions.Action, projectFields map[string]graphql.Fields) error {
	// sort projects to make results stable
	projects := maps.Keys(projectFields)
	slices.Sort(projects)

	for _, project := range projects {
		if size := projectFields[project]["Size"]; size != "" {
			return fmt.Errorf("PR for project %s has size %s", project, size)
		}
	}

	return nil
}

// checkSprint checks if PR is in for current sprint.
func checkSprint(_ *githubactions.Action, projectFields map[string]graphql.Fields) error {
	// sort projects to make results stable
	projects := maps.Keys(projectFields)
	slices.Sort(projects)

	for _, project := range projects {
		sprint, ok := projectFields[project]["Sprint"]
		if !ok {
			continue
		}
		if projectFields[project]["Sprint/IsCurrent"] == "N" {
			return fmt.Errorf("PR for project %s is for sprint %s", project, sprint)
		}
	}

	return nil
}

// checkTitle checks if PR's title does not end with dot.
func checkTitle(_ *githubactions.Action, title string) error {
	titleRegexp := regexp.MustCompile("[a-zA-Z0-9`'\"]$")
	if match := titleRegexp.MatchString(title); !match {
		return fmt.Errorf("PR title must end with a latin letter or digit")
	}

	return nil
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func checkBody(action *githubactions.Action, body string) error {
	action.Debugf("checkBody:\n%s", hex.Dump([]byte(body)))

	// it does not seem to be documented, but PR bodies use CRLF instead of LF for line breaks
	body = strings.ReplaceAll(body, "\r\n", "\n")

	// it is allowed to have a completely empty body
	if len(body) == 0 {
		return nil
	}

	bodyRegexp := regexp.MustCompile(".+[.!?](\n)?$")

	// one \n at the end is allowed, but optional
	if match := bodyRegexp.MatchString(body); !match {
		return fmt.Errorf("PR body must end with dot or other punctuation mark")
	}

	return nil
}
