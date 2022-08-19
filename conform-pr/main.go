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
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"

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
	w := tabwriter.NewWriter(io.MultiWriter(&buf, os.Stderr), 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tCheck\tStatus\t\n")
	fmt.Fprintf(w, "\t-----\t------\t\n")

	conform := true

	for _, res := range runChecks(ctx, action, client) {
		status := ":white_check_mark:"
		if res.err != nil {
			status = ":x: " + res.err.Error()
			conform = false
		}

		fmt.Fprintf(w, "\t%s\t%s\t\n", res.check, status)
	}

	w.Flush()
	action.AddStepSummary(buf.String())

	if !conform {
		action.Fatalf("The PR does not conform to the rules.")
	}
}

// checkResult is a result of a single check.
type checkResult struct {
	check string
	err   error
}

// runChecks runs all the checks included into the PR conformance rules.
func runChecks(ctx context.Context, action *githubactions.Action, client *graphql.Client) []checkResult {
	event, err := internal.ReadEvent(action)
	if err != nil {
		action.Errorf("Failed to read event: %s.", err)
	}

	var nodeID string
	switch event := event.(type) {
	case *github.PullRequestEvent:
		nodeID = *event.PullRequest.NodeID
	default:
		action.Fatalf("Unexpected event type: %T.", event)
	}

	pr := client.GetPullRequest(ctx, nodeID)

	// PRs from dependabot are perfect
	if pr.Author == "dependabot" && pr.AuthorBot {
		return nil
	}

	title := checkResult{
		check: "Title",
		err:   checkTitle(action, pr.Title),
	}
	body := checkResult{
		check: "Body",
		err:   checkBody(action, pr.Body),
	}

	return []checkResult{title, body}
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
