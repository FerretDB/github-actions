package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/github-actions/internal"
	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestDetect(t *testing.T) {
	ctx := context.Background()

	t.Run("pull_request/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_self.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			number: 1,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_fork.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "FerretDB",
			repo:   "dance",
			number: 47,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_self.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			number: 1,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_fork.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "FerretDB",
			repo:   "dance",
			number: 47,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
		}
		assert.Equal(t, expected, actual)
	})
}
