package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

// GitHubClient returns GitHub API client with `GITHUB_TOKEN` from environment, if present.
func GitHubClient(ctx context.Context, action *githubactions.Action) *github.Client {
	token := action.Getenv("GITHUB_TOKEN")
	if token == "" {
		action.Debugf("GITHUB_TOKEN is not set")
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

// ReadEvent reads event from GITHUB_EVENT_PATH path.
func ReadEvent(action *githubactions.Action) (any, error) {
	eventPath := action.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	b, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return nil, err
	}

	// Debug level requires `ACTIONS_RUNNER_DEBUG` secret to be set to `true`:
	// https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/enabling-debug-logging
	// Note that `pull_request` events from forks do not have access to secrets,
	// so that line will not be logged in that case.
	action.Debugf("Read event from %s:\n%s", eventPath, string(b))

	eventName := action.Getenv("GITHUB_EVENT_NAME")
	if eventName == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_NAME is not set")
	}

	var event any
	switch eventName {
	case "pull_request", "pull_request_target":
		event = new(github.PullRequestEvent)
	case "push", "schedule":
		event = new(github.PushEvent)
	default:
		return nil, fmt.Errorf("unhandled event to unmarshal: %q", eventName)
	}

	if err := json.Unmarshal(b, event); err != nil {
		return nil, err
	}

	return event, nil
}
