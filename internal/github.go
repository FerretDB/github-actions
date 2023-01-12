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

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v49/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

// GitHubClient returns GitHub API client with token from the given environment variable.
func GitHubClient(ctx context.Context, action *githubactions.Action, tokenVar string) *github.Client {
	// without the token, our anonymous requests hit the rate limit too often
	token := action.Getenv(tokenVar)
	if token == "" {
		action.Fatalf("%s is not set.", tokenVar)
		return nil
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	httpClient.Transport = NewTransport(httpClient.Transport, action)

	c := github.NewClient(httpClient)
	c.UserAgent = "github-actions/1.0 (+https://github.com/FerretDB/github-actions)"

	// Query authenticated user to check that the client is able to make queries.
	// See https://docs.github.com/en/rest/rate-limit
	// and https://docs.github.com/en/rest/users/users#get-the-authenticated-user.
	user, resp, err := c.Users.Get(ctx, "")
	if err != nil {
		action.Fatalf("Failed to query authenticated user: %s.", err)
		return nil
	}

	action.Infof(
		"User: %s, rate limit: %d/%d, resets at: %s.",
		*user.Login, resp.Rate.Remaining, resp.Rate.Limit, resp.Rate.Reset.Format(time.RFC3339),
	)

	return c
}

// ReadEvent reads event from GITHUB_EVENT_PATH path.
func ReadEvent(action *githubactions.Action) (any, error) {
	eventPath := action.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	b, err := os.ReadFile(eventPath)
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
