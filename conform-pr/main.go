package main

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/FerretDB/github-actions/internal/gh"

	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DebugEnv(action)

	errors := runChecks(action)

	if len(errors) == 0 {
		return
	}

	var serrors []string
	for _, err := range errors {
		serrors = append(serrors, fmt.Sprintf("%s", err))
	}

	action.Fatalf("The PR does not conform to the rules:\n - %s.", strings.Join(serrors, ";\n - "))
}

// runChecks runs all the checks included into the PR conformance rules.
// It returns the list of errors that occurred during the checks.
func runChecks(action *githubactions.Action) []error {
	var errors []error

	pr, err := getPR(action)
	if err != nil {
		return []error{fmt.Errorf("runChecks: %w", err)}
	}

	// PRs from dependabot are perfect
	if pr.author == "dependabot[bot]" {
		return nil
	}

	if err := pr.checkTitle(); err != nil {
		errors = append(errors, err)
	}

	if err := pr.checkBody(); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// getPR returns PR's information.
// If an error occurs, it returns nil and the error.
func getPR(action *githubactions.Action) (*pullRequest, error) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return nil, fmt.Errorf("getPR: %w", err)
	}

	var pr pullRequest
	switch event := event.(type) {
	case *github.PullRequestEvent:
		pr.author = event.PullRequest.User.GetLogin()
		pr.title = event.PullRequest.GetTitle()
		pr.body = event.PullRequest.GetBody()
		pr.nodeID = event.PullRequest.GetNodeID()

		sprints, err := getSprints(action, pr.nodeID)
		if err != nil {
			return nil, fmt.Errorf("getPR: %w", err)
		}
		pr.sprints = sprints
		action.Debugf("PR sprints: %s", sprints)
	default:
		return nil, fmt.Errorf("getPR: unhandled event type %T (only PR-related events are handled)", event)
	}

	return &pr, nil
}

func getSprints(action *githubactions.Action, nodeID string) ([]string, error) {
	ctx := context.Background()
	client := gh.GraphQLClient(ctx, action)
	projects, err := gh.GetPRProjects(client, nodeID)
	if err != nil {
		return nil, fmt.Errorf("getSprints: %w", err)
	}

	var sprints []string
	for _, project := range projects {
		for _, sprint := range project.CurrentSprints {
			sprints = append(sprints, sprint.Title)
		}
	}

	return sprints, nil
}

// pullRequest contains information about PR that is interesting for us.
type pullRequest struct {
	author  string
	title   string
	body    string
	nodeID  string
	sprints []string
}

// checkTitle checks if PR's title does not end with dot.
func (pr *pullRequest) checkTitle() error {
	match, err := regexp.MatchString("[a-zA-Z0-9`'\"]$", pr.title)
	if err != nil {
		return fmt.Errorf("checkTitle: %w", err)
	}

	if !match {
		return fmt.Errorf("checkTitle: PR title must end with a latin letter or digit, but it does not")
	}

	return nil
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func (pr *pullRequest) checkBody() error {
	// it is allowed to have an empty body
	if len(pr.body) == 0 {
		return nil
	}

	match, err := regexp.MatchString(`.+[.!?]$`, pr.body)
	if err != nil {
		return fmt.Errorf("checkBody: %w", err)
	}

	if !match {
		return fmt.Errorf("checkBody: PR body must end with dot or other punctuation mark, but it does not")
	}

	return nil
}
