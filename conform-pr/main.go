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

	if err := checkTitle(action); err != nil {
		action.Fatalf("%s", err)
	}
}

// checkTitle checks if PR's title does not end with dot.
func checkTitle(action *githubactions.Action) error {
	title, err := getPRTitle(action)
	if err != nil {
		return fmt.Errorf("checkTitle: %w", err)
	}

	if strings.HasSuffix(title, ".") {
		return fmt.Errorf("checkTitle: PR title ends with dot")
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

	action.Infof("Got title %q for PR %s", title, url)

	return title, nil
}
