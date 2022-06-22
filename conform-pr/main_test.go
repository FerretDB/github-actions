package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/github-actions/internal/testutil"
)

// stubQuerier implements the simplest gh.Querier interface for testing purposes.
type stubQuerier struct{}

// Query implements gh.Querier interface.
func (sq stubQuerier) Query(context.Context, interface{}, map[string]interface{}) error {
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
		errors := runChecks(action, client)

		assert.Len(t, errors, 0)
	})

	t.Run("pull_request/title_with_dot_body_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_with_dot_body_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action, client)

		// Expect to receive two errors - one for title and one for body
		assert.Len(t, errors, 2)
	})

	t.Run("pull_request/title_without_dot_empty_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action, client)

		assert.Len(t, errors, 0)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action, client)

		assert.Len(t, errors, 0)
	})

	t.Run("pull_request/not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action, client)
		assert.Len(t, errors, 1)
		assert.EqualError(t, errors[0], "runChecks: getPR: unhandled event type *github.PushEvent (only PR-related events are handled)")
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
		assert.EqualError(t, err, "getPR: unhandled event type *github.PushEvent (only PR-related events are handled)")
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
		expectedErr: errors.New("checkTitle: PR title must end with a latin letter or digit, but it does not"),
	}, {
		name:        "pull_request/title_with_whitespace",
		title:       "I'm a title with a whitespace ",
		expectedErr: errors.New("checkTitle: PR title must end with a latin letter or digit, but it does not"),
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
	cases := []struct {
		name        string
		body        string
		expectedErr error
	}{{
		name:        "pull_request/empty_body",
		body:        "",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_dot",
		body:        "I'm a body with a dot.",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_!",
		body:        "I'm a body with a punctuation mark!",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_?",
		body:        "Am I a body with a punctuation mark?",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_without_dot",
		body:        "I'm a body without a dot",
		expectedErr: errors.New("checkBody: PR body must end with dot or other punctuation mark, but it does not"),
	}, {
		name:        "pull_request/body_too_shot",
		body:        "!",
		expectedErr: errors.New("checkBody: PR body must end with dot or other punctuation mark, but it does not"),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pr := pullRequest{
				body: tc.body,
			}
			err := pr.checkBody()
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
