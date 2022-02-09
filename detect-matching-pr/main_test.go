package main

import (
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	t.Run("pull_request/self", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_ACTION_PATH":       "/home/runner/work/_actions/FerretDB/github-actions/detect-matching-pr/detect-matching-pr",
			"GITHUB_ACTION_REF":        "",
			"GITHUB_ACTION_REPOSITORY": "",
			"GITHUB_ACTION":            "__FerretDB_github-actions",
			"GITHUB_ACTIONS":           "true",
			"GITHUB_ACTOR":             "AlekSi",
			"GITHUB_API_URL":           "https://api.github.com",
			"GITHUB_BASE_REF":          "main",
			"GITHUB_ENV":               "/home/runner/work/_temp/_runner_file_commands/set_env_a9ad5c5f-b89c-4e5a-9ae3-c493ded55189",
			"GITHUB_EVENT_NAME":        "pull_request",
			"GITHUB_EVENT_PATH":        "/home/runner/work/_temp/_github_workflow/event.json",
			"GITHUB_GRAPHQL_URL":       "https://api.github.com/graphql",
			"GITHUB_HEAD_REF":          "feature-branch",
			"GITHUB_JOB":               "detect-matching-pr",
			"GITHUB_PATH":              "/home/runner/work/_temp/_runner_file_commands/add_path_a9ad5c5f-b89c-4e5a-9ae3-c493ded55189",
			"GITHUB_REF_NAME":          "10/merge",
			"GITHUB_REF_PROTECTED":     "false",
			"GITHUB_REF_TYPE":          "branch",
			"GITHUB_REF":               "refs/pull/10/merge",
			"GITHUB_REPOSITORY_OWNER":  "AlekSi",
			"GITHUB_REPOSITORY":        "AlekSi/FerretDB",
			"GITHUB_RETENTION_DAYS":    "90",
			"GITHUB_RUN_ATTEMPT":       "1",
			"GITHUB_RUN_ID":            "1795408545",
			"GITHUB_RUN_NUMBER":        "74",
			"GITHUB_SERVER_URL":        "https://github.com",
			"GITHUB_SHA":               "88db08f7839f77424263de0137415180f68fe3a3",
			"GITHUB_WORKFLOW":          "Go",
			"GITHUB_WORKSPACE":         "/home/runner/work/FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := detect(action)
		require.NoError(t, err)
		assert.Equal(t, "AlekSi", result.owner)
	})

	t.Run("pull_request/fork", func(t *testing.T) {
		getEnv := getEnvFunc(t, map[string]string{
			"GITHUB_ACTION_PATH":       "/home/runner/work/_actions/FerretDB/github-actions/detect-matching-pr/detect-matching-pr",
			"GITHUB_ACTION_REF":        "",
			"GITHUB_ACTION_REPOSITORY": "",
			"GITHUB_ACTION":            "__FerretDB_github-actions",
			"GITHUB_ACTIONS":           "true",
			"GITHUB_ACTOR":             "AlekSi",
			"GITHUB_API_URL":           "https://api.github.com",
			"GITHUB_BASE_REF":          "main",
			"GITHUB_ENV":               "/home/runner/work/_temp/_runner_file_commands/set_env_a0920f24-4ead-48e1-9424-4e926f05a72f",
			"GITHUB_EVENT_NAME":        "pull_request",
			"GITHUB_EVENT_PATH":        "/home/runner/work/_temp/_github_workflow/event.json",
			"GITHUB_GRAPHQL_URL":       "https://api.github.com/graphql",
			"GITHUB_HEAD_REF":          "feature-branch",
			"GITHUB_JOB":               "detect-matching-pr",
			"GITHUB_PATH":              "/home/runner/work/_temp/_runner_file_commands/add_path_a0920f24-4ead-48e1-9424-4e926f05a72f",
			"GITHUB_REF_NAME":          "305/merge",
			"GITHUB_REF_PROTECTED":     "false",
			"GITHUB_REF_TYPE":          "branch",
			"GITHUB_REF":               "refs/pull/305/merge",
			"GITHUB_REPOSITORY_OWNER":  "FerretDB",
			"GITHUB_REPOSITORY":        "FerretDB/FerretDB",
			"GITHUB_RETENTION_DAYS":    "90",
			"GITHUB_RUN_ATTEMPT":       "1",
			"GITHUB_RUN_ID":            "1795410183",
			"GITHUB_RUN_NUMBER":        "781",
			"GITHUB_SERVER_URL":        "https://github.com",
			"GITHUB_SHA":               "057ee53daee782c81944406dae03fbd6214f91a3",
			"GITHUB_WORKFLOW":          "Go",
			"GITHUB_WORKSPACE":         "/home/runner/work/FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := detect(action)
		require.NoError(t, err)
		assert.Equal(t, "AlekSi", result.owner)
	})
}
