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
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/github-actions/internal/testutil"
)

// stubQuerier implements the simplest graphql.Querier interface for testing purposes.
type stubQuerier struct{}

// Query implements graphql.Querier interface.
func (sq stubQuerier) Query(context.Context, any, map[string]any) error {
	return nil
}

func TestRunChecks(t *testing.T) {
	client := stubQuerier{}

	t.Run("pull_request/title_without_dot_body_with_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action, client)

		expectedSummaries := []Summary{{"Title", nil}, {"Body", nil}}
		assert.Equal(t, expectedSummaries, summaries, 2)
	})

	t.Run("pull_request/title_with_dot_body_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_with_dot_body_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action, client)

		expectedSummaries := []Summary{
			{"Title", fmt.Errorf("PR title must end with a latin letter or digit")},
			{"Body", fmt.Errorf("PR body must end with dot or other punctuation mark")},
		}
		assert.Equal(t, expectedSummaries, summaries, 2)
	})

	t.Run("pull_request/title_without_dot_empty_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action, client)

		expectedSummaries := []Summary{{"Title", nil}, {"Body", nil}}
		assert.Equal(t, expectedSummaries, summaries, 2)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action, client)

		assert.Len(t, summaries, 0)
	})

	t.Run("pull_request/not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action, client)
		assert.Len(t, summaries, 1)
		assert.EqualError(t, summaries[0].Error, "unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}

func TestGetPR(t *testing.T) {
	client := stubQuerier{}

	t.Run("pull_request/with_title_and_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		pr, err := getPR(action, client)
		assert.NoError(t, err)
		assert.Equal(t, "Add Docker badge", pr.title)
		assert.Equal(t, "This PR is a sample PR \n\nrepresenting a body that ends with a dot.", pr.body)
	})

	t.Run("pull_request/title_without_dot_empty_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		pr, err := getPR(action, client)
		assert.NoError(t, err)
		assert.Equal(t, "Add Docker badge", pr.title)
		assert.Empty(t, pr.body)
	})

	t.Run("pull_request/not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		pr, err := getPR(action, client)
		assert.Nil(t, pr)
		assert.EqualError(t, err, "unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}

func TestCheckTitle(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		expectedErr error
	}{{
		name:        "pull_request/title_without_dot",
		title:       "I'm a title without a dot",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_a_digit",
		title:       "I'm a title without a digit 1",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_dot",
		title:       "I'm a title with a dot.",
		expectedErr: errors.New("PR title must end with a latin letter or digit"),
	}, {
		name:        "pull_request/title_with_whitespace",
		title:       "I'm a title with a whitespace ",
		expectedErr: errors.New("PR title must end with a latin letter or digit"),
	}, {
		name:        "pull_request/title_with_backticks",
		title:       "I'm a title with a `backticks`",
		expectedErr: nil,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pr := pullRequest{
				title: tc.title,
			}
			err := pr.checkTitle()
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestCheckBody(t *testing.T) {
	errNoPunctuation := errors.New("PR body must end with dot or other punctuation mark")

	cases := []struct {
		name        string
		body        string
		expectedErr error
	}{{
		name:        "pull_request/empty_body",
		body:        "",
		expectedErr: nil,
	}, {
		name:        "pull_request/whitespace_body",
		body:        "\n",
		expectedErr: errNoPunctuation,
	}, {
		name:        "pull_request/body_with_dot",
		body:        "I'm a body with a dot.",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_!",
		body:        "I'm a body with a punctuation mark!\r\n",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_?",
		body:        "Am I a body with a punctuation mark?",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_without_dot",
		body:        "I'm a body without a dot\n",
		expectedErr: errNoPunctuation,
	}, {
		name:        "pull_request/body_too_shot",
		body:        "!\r\n",
		expectedErr: errNoPunctuation,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pr := pullRequest{
				body: tc.body,
			}
			action := githubactions.New()
			err := pr.checkBody(action)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
