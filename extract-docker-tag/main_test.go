package main

import (
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestExtract(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/pull/1/merge",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "pr-extract-docker-tag", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "pr-extract-docker-tag", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
			"GITHUB_REF_NAME":   "58/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/pull/58/merge",
			"GITHUB_REPOSITORY": "FerretDB/dance",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "pr-mongo-go-driver-29d768e", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("push/main", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "main", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("schedule", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "main", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "main", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("push/semVerTag", func(t *testing.T) {
		// v[0-9].[0-9]+.[0-9]+
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REF":        "refs/tags/v0.0.1",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "0.0.1", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})

	t.Run("push/tagWrong", func(t *testing.T) {
		// v[0-9].[0-9]+.[0-9]+
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REF":        "refs/tags/v0.0.",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := extract(action)
		require.NoError(t, err)
		assert.Equal(t, "ferretdb", result.owner)
		assert.Equal(t, "ferretdb-dev", result.name)
		assert.Equal(t, "main", result.tag)
		assert.Equal(t, "ghcr.io/ferretdb/ferretdb-dev", result.ghcr)
	})
}
