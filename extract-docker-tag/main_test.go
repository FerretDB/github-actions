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

func TestExtract(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":    "main",
			"GITHUB_EVENT_NAME":  "pull_request",
			"GITHUB_HEAD_REF":    "extract-docker-tag",
			"GITHUB_REF_NAME":    "1/merge",
			"GITHUB_REF_TYPE":    "branch",
			"GITHUB_REF":         "refs/pull/1/merge",
			"GITHUB_REPOSITORY":  "FerretDB/github-actions",
			"GITHUB_RUN_ATTEMPT": "1",
			"GITHUB_RUN_ID":      "1634171996",
			"GITHUB_RUN_NUMBER":  "2",
			"GITHUB_SHA":         "a62653e4dcb5b0f5100a4466a2455e91e68f37b3",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		owner, name, tag, err := extract(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", owner)
		assert.Equal(t, "github-actions", name)
		assert.Equal(t, "dev-extract-docker-tag", tag)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/github-actions",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		owner, name, tag, err := extract(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", owner)
		assert.Equal(t, "github-actions", name)
		assert.Equal(t, "dev-extract-docker-tag", tag)
	})

	t.Run("push/main", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/github-actions",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		owner, name, tag, err := extract(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", owner)
		assert.Equal(t, "github-actions", name)
		assert.Equal(t, "main", tag)
	})

	t.Run("schedule", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/github-actions",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		owner, name, tag, err := extract(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", owner)
		assert.Equal(t, "github-actions", name)
		assert.Equal(t, "main", tag)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/github-actions",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		owner, name, tag, err := extract(action, getEnv)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", owner)
		assert.Equal(t, "github-actions", name)
		assert.Equal(t, "main", tag)
	})
}
