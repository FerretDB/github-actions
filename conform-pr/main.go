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
	action.AddStepSummary("|----------------|-----------------------------------------|")

	for _, summary := range summaries {
		statusSign := ":heavy_multiplication_x:"
		if summary.Ok {
			statusSign = ":heavy_check_mark:"
		}
		action.AddStepSummary(fmt.Sprintf("|%s | %s %s|", summary.Name, statusSign, summary.Details))
	}
	action.AddStepSummary("|--------|")

	for _, v := range summaries {
		if !v.Ok {
			action.Fatalf("The PR does not conform to the rules")
		}
	}
}

type Summary struct {
	Name    string
	Ok      bool
	Details error
}

// runChecks runs all the checks included into the PR conformance rules.
// It returns the list of check summary for the checks.
func runChecks(action *githubactions.Action, client graphql.Querier) []Summary {
	pr, summaries := getPR(action, client)
	if len(summaries) > 0 {
		return summaries
	}

	// PRs from dependabot are perfect
	if pr.author == "dependabot[bot]" {
		return nil
	}

	summaries = pr.checkTitle()

	summaries = append(summaries, pr.checkBody(action)...)

	return summaries
}

// getPR returns PR's information and returns
// * pull request details if no errors occured
// * a summary list whether and which check passed successfully or not.
func getPR(action *githubactions.Action, client graphql.Querier) (*pullRequest, []Summary) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return nil, []Summary{{Name: "Read event", Details: err}}
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
			return nil, []Summary{{
				Name:    "Node fields",
				Details: err,
			}}
		}
		pr.values = values
		action.Infof("getPR: Values: %v", values)
	default:
		return nil, []Summary{{
			Name:    "Event type",
			Details: fmt.Errorf("unhandled event type %T (only PR-related events are handled)", event),
		}}
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

// checkTitle checks if PR's title does not end with dot and returns a summary list for checks.
func (pr *pullRequest) checkTitle() []Summary {
	match, err := regexp.MatchString("[a-zA-Z0-9`'\"]$", pr.title)
	if err != nil {
		return []Summary{{Name: "Title regex parsing", Details: err}}
	}

	titleMatches := Summary{Name: "PR title must end with a latin letter or digit"}
	if match {
		titleMatches.Ok = true
	}
	return []Summary{titleMatches}
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func (pr *pullRequest) checkBody(action *githubactions.Action) []Summary {
	action.Debugf("checkBody:\n%s", hex.Dump([]byte(pr.body)))

	// it does not seem to be documented, but PR bodies use CRLF instead of LF for line breaks
	pr.body = strings.ReplaceAll(pr.body, "\r\n", "\n")

	// it is allowed to have a completely empty body
	if len(pr.body) == 0 {
		return nil
	}

	// one \n at the end is allowed, but optional
	match, err := regexp.MatchString(".+[.!?](\n)?$", pr.body)
	if err != nil {
		return []Summary{{Name: "Body regex parsing", Details: err}}
	}

	bodyCheck := Summary{Name: "PR body must end with dot or other punctuation mark"}
	if match {
		bodyCheck.Ok = true
	}
	return []Summary{bodyCheck}
}
