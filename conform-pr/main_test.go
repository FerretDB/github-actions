package main

import (
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestRunChecks(t *testing.T) {
	t.Run("pull_request/title_without_dot_body_with_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action)

		assert.Len(t, summaries, 2)
	})

	t.Run("pull_request/title_with_dot_body_without_dot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_with_dot_body_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action)

		// Expect to receive two errors - one for title and one for body
		assert.Len(t, summaries, 2)
	})

	t.Run("pull_request/title_without_dot_empty_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_title_without_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action)

		assert.Len(t, summaries, 1)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action)

		assert.Len(t, summaries, 0)
	})

	t.Run("pull_request/not_a_pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		summaries := runChecks(action)
		assert.Len(t, summaries, 1)
		assert.EqualError(t, summaries[0].Details, "unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}

func TestGetPR(t *testing.T) {
	t.Run("pull_request/with_title_and_body", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_body_with_dot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		pr, summaries := getPR(action)
		assert.Len(t, summaries, 0)
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
		pr, summaries := getPR(action)
		assert.Len(t, summaries, 0)
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
		pr, summaries := getPR(action)
		assert.Len(t, summaries, 1)
		assert.Nil(t, pr)
		assert.EqualError(t, summaries[0].Details, "unhandled event type *github.PushEvent (only PR-related events are handled)")
	})
}

func TestCheckTitle(t *testing.T) {
	cases := []struct {
		name              string
		title             string
		expectedSummaries []Summary
	}{
		{
			name:              "pull_request/title_without_dot",
			title:             "I'm a title without a dot",
			expectedSummaries: []Summary{{Name: "PR title must end with a latin letter or digit", Ok: true}},
		},
		{
			name:              "pull_request/title_with_a_number",
			title:             "I'm a title that ends with a number",
			expectedSummaries: []Summary{{Name: "PR title must end with a latin letter or digit", Ok: true}},
		},
		{
			name:              "pull_request/title_with_dot",
			title:             "I'm a title with a dot.",
			expectedSummaries: []Summary{{Name: "PR title must end with a latin letter or digit", Ok: false}},
		},
		{
			name:              "pull_request/title_with_whitespace",
			title:             "I'm a title with a whitespace ",
			expectedSummaries: []Summary{{Name: "PR title must end with a latin letter or digit", Ok: false}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pr := pullRequest{
				title: tc.title,
			}
			summaries := pr.checkTitle()
			assert.Len(t, summaries, 1)
			assert.Equal(t, tc.expectedSummaries, summaries)
		})
	}
}

func TestCheckBody(t *testing.T) {
	cases := []struct {
		name              string
		body              string
		expectedSummaries []Summary
	}{
		{
			name:              "pull_request/empty_body",
			body:              "",
			expectedSummaries: nil,
		},
		{
			name:              "pull_request/body_with_dot",
			body:              "I'm a body with a dot.",
			expectedSummaries: []Summary{{Name: "PR body must end with dot or other punctuation mark", Ok: true}},
		},
		{
			name:              "pull_request/body_with_!",
			body:              "I'm a body with a punctuation mark!",
			expectedSummaries: []Summary{{Name: "PR body must end with dot or other punctuation mark", Ok: true}},
		},
		{
			name:              "pull_request/body_with_?",
			body:              "Am I a body with a punctuation mark?",
			expectedSummaries: []Summary{{Name: "PR body must end with dot or other punctuation mark", Ok: true}},
		},
		{
			name:              "pull_request/body_without_dot",
			body:              "I'm a body without a dot",
			expectedSummaries: []Summary{{Name: "PR body must end with dot or other punctuation mark", Ok: false}},
		},
		{
			name:              "pull_request/body_too_short",
			body:              "!",
			expectedSummaries: []Summary{{Name: "PR body must end with dot or other punctuation mark", Ok: false}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pr := pullRequest{
				body: tc.body,
			}
			err := pr.checkBody()
			assert.Equal(t, tc.expectedSummaries, err)
		})
	}
}
