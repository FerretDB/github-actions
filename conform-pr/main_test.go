package main

import (
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestRunChecks(t *testing.T) {
	t.Run("pull_request/body_with_dot_title_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action)
		assert.Len(t, errors, 0)
	})

	t.Run("pull_request/body_without_dot_title_with_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_without_dot_title_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		errors := runChecks(action)

		// Expect to receive two errors - one for title and one for body
		assert.Len(t, errors, 2)
	})
}

func TestCheckTitle(t *testing.T) {
	t.Run("pull_request/title_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkTitle(action)
		assert.NoError(t, err)
	})

	t.Run("pull_request/title_with_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkTitle(action)
		assert.EqualError(t, err, "checkTitle: PR title must not end with dot, but it does")
	})

	t.Run("not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkTitle(action)
		assert.EqualError(t, err, "checkTitle: getPRTitle: unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}

func TestCheckBody(t *testing.T) {
	t.Run("pull_request/body_with_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkBody(action)
		assert.NoError(t, err)
	})

	t.Run("pull_request/body_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_without_dot_title_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkBody(action)
		assert.EqualError(t, err, "checkBody: PR body must end with dot, but it does not")
	})

	t.Run("not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		err := checkBody(action)
		assert.EqualError(t, err, "checkBody: getPRBody: unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}
