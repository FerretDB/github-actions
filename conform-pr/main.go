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
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/google/go-github/v49/github"
	"github.com/jdkato/prose/v2"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/github-actions/internal"
	"github.com/FerretDB/github-actions/internal/graphql"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	action := githubactions.New()
	client := internal.GitHubClient(ctx, action, "GITHUB_TOKEN")
	gClient := graphql.NewClient(ctx, action, "CONFORM_TOKEN")

	internal.DebugEnv(action)

	event, err := internal.ReadEvent(action)
	if err != nil {
		action.Errorf("Failed to read event: %s.", err)
	}

	prEvent, ok := event.(*github.PullRequestEvent)
	if !ok {
		action.Fatalf("Unexpected event type: %T.", event)
	}

	c := &checker{
		action:  action,
		client:  client,
		gClient: gClient,
	}

	results, community := c.runChecks(
		ctx,
		*prEvent.Organization.Login, *prEvent.PullRequest.User.Login, *prEvent.PullRequest.NodeID,
	)

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tCheck\tStatus\t\n")
	fmt.Fprintf(w, "\t-----\t------\t\n")

	conform := true
	for _, res := range results {
		status := "✅"
		if res.err != nil {
			status = "❌ " + res.err.Error()
			conform = false
		}

		fmt.Fprintf(w, "\t%s\t%s\t\n", res.check, status)
	}

	w.Flush()
	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	if !conform {
		if community {
			action.Fatalf("Maintainers will update that PR to conform to the project's standards.")
		}

		action.Fatalf("PR does not conform to the project's standards.")
	}
}

// checker holds state shared by all checks.
type checker struct {
	action  *githubactions.Action
	client  *github.Client
	gClient *graphql.Client
}

// checkResult is a result of a single check.
type checkResult struct {
	check string
	err   error
}

// runChecks runs all the checks for the given PR.
//
// It returns check results and a flag indicating if the PR is from the community (true if yet).
func (c *checker) runChecks(ctx context.Context, org, user, nodeID string) ([]checkResult, bool) {
	members, _, err := c.client.Organizations.ListMembers(ctx, org, &github.ListMembersOptions{
		PublicOnly: true,
	})
	if err != nil {
		c.action.Fatalf("Failed to list organization members: %s.", err)
	}

	community := true

	for _, m := range members {
		if *m.Login == user {
			community = false
			break
		}
	}

	pr := c.gClient.GetPullRequest(ctx, nodeID)

	// Be nice to dependabot's compatibility scoring feature:
	// https://docs.github.com/en/code-security/dependabot/dependabot-security-updates/about-dependabot-security-updates#about-compatibility-scores
	//nolint:lll // that URL is long
	if pr.Author == "dependabot" && pr.AuthorBot {
		return nil, community
	}

	var res []checkResult

	res = append(res, checkResult{
		check: "Labels",
		err:   checkLabels(c.action, pr.Labels),
	})
	res = append(res, checkResult{
		check: "Size",
		err:   checkSize(c.action, pr.ProjectFields),
	})
	res = append(res, checkResult{
		check: "Sprint",
		err:   checkSprint(c.action, pr.ProjectFields, community),
	})
	res = append(res, checkResult{
		check: "Title",
		err:   checkTitle(c.action, pr.Title),
	})
	res = append(res, checkResult{
		check: "Body",
		err:   checkBody(c.action, pr.Body),
	})
	res = append(res, checkResult{
		check: "Auto-merge",
		err:   checkAutoMerge(c.action, pr, community),
	})

	return res, community
}

// checkLabels checks if PR's labels are valid.
func checkLabels(_ *githubactions.Action, labels []string) error {
	var res []string

	for _, l := range []string{
		"good first issue",
		"help wanted",
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
		return fmt.Errorf("Those labels should not be applied to PRs: %s.", strings.Join(res, ", "))
	}

	if slices.Contains(labels, "do not merge") {
		return fmt.Errorf("That PR should not be merged yet.")
	}

	if slices.Contains(labels, "not ready") {
		return fmt.Errorf("That PR can't be merged yet; remove `not ready` label.")
	}

	return nil
}

// checkSize checks that PR has a "Size" field unset.
func checkSize(_ *githubactions.Action, projectFields map[string]graphql.Fields) error {
	// sort projects to make results stable
	projects := maps.Keys(projectFields)
	slices.Sort(projects)

	var got []string
	for _, project := range projects {
		if size := projectFields[project]["Size"]; size != "" {
			got = append(got, fmt.Sprintf("%q for project %q", size, project))
		}
	}

	if got != nil {
		return fmt.Errorf(`PR should have "Size" field unset, got %s.`, strings.Join(got, ", "))
	}

	return nil
}

// checkSprint checks that PR has a "Sprint" field set.
func checkSprint(_ *githubactions.Action, projectFields map[string]graphql.Fields, community bool) error {
	// sort projects to make results stable
	projects := maps.Keys(projectFields)
	slices.Sort(projects)

	for _, project := range projects {
		if sprint := projectFields[project]["Sprint"]; sprint != "" {
			return nil
		}
	}

	msg := `PR should have "Sprint" field set.`
	if community {
		msg += ` Don't worry, maintainers will set it for you.`
	}

	return errors.New(msg)
}

// checkTitle checks if PR's title does not end with dot
// and also check if it starts with an imperative verb.
func checkTitle(_ *githubactions.Action, title string) error {
	uppercaseRegexp := regexp.MustCompile("^[A-Z]+")
	if match := uppercaseRegexp.MatchString(title); !match {
		return fmt.Errorf("PR title must start with an uppercase letter.")
	}

	titleRegexp := regexp.MustCompile("[a-zA-Z0-9`'\"]$")
	if match := titleRegexp.MatchString(title); !match {
		return fmt.Errorf("PR title must end with a latin letter or digit.")
	}

	firstWord := strings.Split(title, " ")[0]
	doc, err := prose.NewDocument("I " + firstWord)
	if err != nil {
		return fmt.Errorf("error parsing PR title.")
	}

	tokens := doc.Tokens()
	tok := tokens[1]

	// imperative verbs have "VBP" tag
	// https://github.com/jdkato/prose/tree/v2#tagging
	if tok.Tag != "VBP" {
		return fmt.Errorf("PR title must start with an imperative verb (got %q).", tok.Tag)
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
		return fmt.Errorf("PR body must end with dot or other punctuation mark.")
	}

	return nil
}

// checkAutoMerge checks if PR's auto-merge is enabled.
func checkAutoMerge(_ *githubactions.Action, pr *graphql.PullRequest, community bool) error {
	if pr.Closed || pr.AutoMerge {
		return nil
	}

	msg := `PR should have auto-merge enabled.`
	if community {
		msg += ` Don't worry, maintainers will enable it for you.`
	}

	return errors.New(msg)
}
