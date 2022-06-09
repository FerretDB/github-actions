package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"

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

	title, body, err := getPR(action)
	if err != nil {
		return []error{fmt.Errorf("runChecks: %w", err)}
	}

	if err := checkTitle(title); err != nil {
		errors = append(errors, err)
	}

	if err := checkBody(body); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// checkTitle checks if PR's title does not end with dot.
func checkTitle(title string) error {
	match, err := regexp.MatchString(`[a-zA-Z0-9]$`, title)
	if err != nil {
		return fmt.Errorf("checkTitle: %w", err)
	}

	if !match {
		return fmt.Errorf("checkTitle: PR title must end with a latin letter or digit, but it does not")
	}

	return nil
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func checkBody(body string) error {
	// It is allowed to have empty body.
	if len(body) == 0 {
		return nil
	}

	match, err := regexp.MatchString(`.+[.!?]$`, body)
	if err != nil {
		return fmt.Errorf("checkBody: %w", err)
	}

	if !match {
		return fmt.Errorf("checkBody: PR body must end with dot or other punctuation mark, but it does not")
	}

	return nil
}

// getPR returns PR's title and body.
// If an error occurs, it returns empty strings for title and body and the error.
func getPR(action *githubactions.Action) (title, body string, err error) {
	var event interface{}
	event, err = internal.ReadEvent(action)
	if err != nil {
		return "", "", fmt.Errorf("getPR: %w", err)
	}

	var url string

	switch event := event.(type) {
	case *github.PullRequestEvent:
		title = *event.PullRequest.Title
		if event.PullRequest.Body == nil {
			body = ""
		} else {
			body = *event.PullRequest.Body
		}
		url = *event.PullRequest.URL
	default:
		return "", "", fmt.Errorf("getPR: unhandled event type %T (only PR-related events are handled)", event)
	}

	action.Infof("Got title %q and body %q for PR %s", title, body, url)
	return title, body, nil
}
