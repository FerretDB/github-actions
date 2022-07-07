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

	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
	"github.com/FerretDB/github-actions/internal/graphql"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DebugEnv(action)

	// graphQL client is used to get PR's projects
	ctx := context.Background()
	client, err := graphql.GraphQLClient(ctx, action, "CONFORM_TOKEN")
	if err != nil {
		action.Fatalf("main: %s", err)
	}

	summaries := runChecks(action, client)
	action.AddStepSummary("| Check  | Status |")
	action.AddStepSummary("|--------|--------|")

	for _, summary := range summaries {
		statusSign := ":x:"
		if summary.Error == nil {
			statusSign = ":white_check_mark:"
		}
		if summary.Error != nil {
			action.AddStepSummary(fmt.Sprintf("|%s | %s %s|", summary.Name, statusSign, summary.Error))
		} else {
			action.AddStepSummary(fmt.Sprintf("|%s | %s |", summary.Name, statusSign))
		}
	}

	for _, v := range summaries {
		if v.Error != nil {
			action.Fatalf("The PR does not conform to the rules")
		}
	}
}

// Summary is a markdown summary.
type Summary struct {
	Name  string
	Error error
}

// runChecks runs all the checks included into the PR conformance rules.
// It returns the list of check summary for the checks.
func runChecks(action *githubactions.Action, client graphql.Querier) []Summary {
	pr, err := getPR(action, client)
	if err != nil {
		return []Summary{{Name: "Read PR", Error: err}}
	}

	// PRs from dependabot are perfect
	if pr.author == "dependabot[bot]" {
		return nil
	}

	titleSummary := Summary{Name: "Title"}
	titleSummary.Error = pr.checkTitle()

	bodySummary := Summary{Name: "Body"}
	bodySummary.Error = pr.checkBody(action)

	return []Summary{titleSummary, bodySummary}
}

// getPR returns PR's information.
// If an error occurs, it returns nil and the error.
func getPR(action *githubactions.Action, client graphql.Querier) (*pullRequest, error) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return nil, fmt.Errorf("Read event: %w", err)
	}

	var pr pullRequest
	switch event := event.(type) {
	case *github.PullRequestEvent:
		pr.author = event.PullRequest.User.GetLogin()
		pr.title = event.PullRequest.GetTitle()
		pr.body = event.PullRequest.GetBody()
		pr.nodeID = event.PullRequest.GetNodeID()

		action.Debugf("getPR: Node ID is: %s", pr.nodeID)
		values, err := getFieldValues(client, pr.nodeID)
		if err != nil {
			return nil, fmt.Errorf("Get node fields: %w", err)
		}
		pr.values = values
		action.Infof("getPR: Values: %v", values)
	default:
		return nil, fmt.Errorf("unhandled event type %T (only PR-related events are handled)", event)
	}
	return &pr, nil
}

// getFieldValues returns the list of field values for the given PR node ID.
func getFieldValues(client graphql.Querier, nodeID string) (map[string]string, error) {
	items, err := graphql.GetPRItems(client, nodeID)
	if err != nil {
		return nil, fmt.Errorf("getFieldValues: %w", err)
	}

	values := make(map[string]string)
	for _, item := range items {
		for _, value := range item.FieldValues.Nodes {
			values[string(value.ProjectField.Name)] = value.ValueTitle
		}
	}

	return values, nil
}

// pullRequest contains information about PR that is interesting for us.
type pullRequest struct {
	author string
	title  string
	body   string
	nodeID string
	values map[string]string
}

// checkTitle checks if PR's title does not end with dot.
func (pr *pullRequest) checkTitle() error {
	titleRegexp := regexp.MustCompile("[a-zA-Z0-9`'\"]$")
	if match := titleRegexp.MatchString(pr.title); !match {
		return fmt.Errorf("PR title must end with a latin letter or digit")
	}
	return nil
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func (pr *pullRequest) checkBody(action *githubactions.Action) error {
	action.Debugf("checkBody:\n%s", hex.Dump([]byte(pr.body)))

	// it does not seem to be documented, but PR bodies use CRLF instead of LF for line breaks
	pr.body = strings.ReplaceAll(pr.body, "\r\n", "\n")

	// it is allowed to have a completely empty body
	if len(pr.body) == 0 {
		return nil
	}

	bodyRegexp := regexp.MustCompile(".+[.!?](\n)?$")

	// one \n at the end is allowed, but optional
	if match := bodyRegexp.MatchString(pr.body); !match {
		return fmt.Errorf("PR body must end with dot or other punctuation mark")
	}
	return nil
}
