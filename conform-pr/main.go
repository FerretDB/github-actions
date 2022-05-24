package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/google/go-github/v42/github"
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

	if err := checkTitle(action); err != nil {
		errors = append(errors, err)
	}

	if err := checkBody(action); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// checkTitle checks if PR's title does not end with dot.
func checkTitle(action *githubactions.Action) error {
	title, err := getPRTitle(action)
	if err != nil {
		return fmt.Errorf("checkTitle: %w", err)
	}

	if strings.HasSuffix(title, ".") {
		return fmt.Errorf("checkTitle: PR title must not end with dot, but it does")
	}

	return nil
}

// checkBody checks if PR's body (description) ends with dot.
func checkBody(action *githubactions.Action) error {
	title, err := getPRBody(action)
	if err != nil {
		return fmt.Errorf("checkBody: %w", err)
	}

	if !strings.HasSuffix(title, ".") {
		return fmt.Errorf("checkBody: PR body must end with dot, but it does not")
	}

	return nil
}

// getPRTitle returns PR's title.
func getPRTitle(action *githubactions.Action) (string, error) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return "", fmt.Errorf("getPRTitle: %w", err)
	}

	var title string
	var url string

	switch event := event.(type) {
	case *github.PullRequestEvent:
		title = *event.PullRequest.Title
		url = *event.PullRequest.URL

	default:
		return "", fmt.Errorf("getPRTitle: unhandled event type %T (only PR-related events are handled)", event)
	}

	action.Infof("Got title `%q` for PR %s", title, url)

	return title, nil
}

// getPRBody returns PR's description.
func getPRBody(action *githubactions.Action) (string, error) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return "", fmt.Errorf("getPRBody: %w", err)
	}

	var body string
	var url string

	switch event := event.(type) {
	case *github.PullRequestEvent:
		body = *event.PullRequest.Body
		url = *event.PullRequest.URL

	default:
		return "", fmt.Errorf("getPRBody: unhandled event type %T (only PR-related events are handled)", event)
	}

	action.Infof("Got body `%q` for PR %s", body, url)

	return body, nil
}
