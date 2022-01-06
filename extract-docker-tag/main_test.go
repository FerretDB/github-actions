package main

import (
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getEnvFunc(t *testing.T, env map[string]string) func(string) string {
	return func(key string) string {
		if val, ok := env[key]; ok {
			return val
		}

		t.Fatalf("unexpected key %q", key)
		panic("not reached")
	}
}

func TestExtractDockerTag(t *testing.T) {
	t.Run("PullRequest", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":    "main",
			"GITHUB_EVENT_NAME":  "pull_request",
			"GITHUB_HEAD_REF":    "extract-docker-tag",
			"GITHUB_REF_NAME":    "1/merge",
			"GITHUB_REF_TYPE":    "branch",
			"GITHUB_REF":         "refs/pull/1/merge",
			"GITHUB_RUN_ATTEMPT": "1",
			"GITHUB_RUN_ID":      "1634171996",
			"GITHUB_RUN_NUMBER":  "2",
			"GITHUB_SHA":         "a62653e4dcb5b0f5100a4466a2455e91e68f37b3",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extractDockerTag(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "dev-extract-docker-tag", actual)
	})

	t.Run("PushMain", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":    "",
			"GITHUB_EVENT_NAME":  "push",
			"GITHUB_HEAD_REF":    "",
			"GITHUB_REF_NAME":    "main",
			"GITHUB_REF_TYPE":    "branch",
			"GITHUB_REF":         "refs/heads/main",
			"GITHUB_RUN_ATTEMPT": "1",
			"GITHUB_RUN_ID":      "1634463356",
			"GITHUB_RUN_NUMBER":  "10",
			"GITHUB_SHA":         "82c9d4f3b8d63fb67c0661c447ba0a2eef98ab35",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extractDockerTag(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "main", actual)
	})

	t.Run("Cron", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":    "",
			"GITHUB_EVENT_NAME":  "schedule",
			"GITHUB_HEAD_REF":    "",
			"GITHUB_REF_NAME":    "main",
			"GITHUB_REF_TYPE":    "branch",
			"GITHUB_REF":         "refs/heads/main",
			"GITHUB_RUN_ATTEMPT": "1",
			"GITHUB_RUN_ID":      "1661466966",
			"GITHUB_RUN_NUMBER":  "20",
			"GITHUB_SHA":         "2065c6bd726887f84869835180d033c82d39b3d4",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extractDockerTag(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "main", actual)
	})
}
