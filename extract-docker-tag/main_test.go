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
			"GITHUB_BASE_REF":         "main",
			"GITHUB_EVENT_NAME":       "pull_request",
			"GITHUB_HEAD_REF":         "extract-docker-tag",
			"GITHUB_JOB":              "test",
			"GITHUB_REF_NAME":         "1/merge",
			"GITHUB_REF_PROTECTED":    "false",
			"GITHUB_REF_TYPE":         "branch",
			"GITHUB_REF":              "refs/pull/1/merge",
			"GITHUB_REPOSITORY_OWNER": "FerretDB",
			"GITHUB_REPOSITORY":       "FerretDB/github-actions",
			"GITHUB_RUN_ATTEMPT":      "1",
			"GITHUB_RUN_ID":           "1634171996",
			"GITHUB_RUN_NUMBER":       "2",
			"GITHUB_SHA":              "a62653e4dcb5b0f5100a4466a2455e91e68f37b3",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extractDockerTag(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "extract-docker-tag", actual)
	})
}
