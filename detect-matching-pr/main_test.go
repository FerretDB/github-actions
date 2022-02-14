package main

import (
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestDetect(t *testing.T) {
	t.Run("pull_request/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_self.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "feature-branch",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_fork.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "FerretDB",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "feature-branch",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_dependabot.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "dependabot/go_modules/tools/github.com/reviewdog/reviewdog-0.14.0",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_self.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "feature-branch",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_fork.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "FerretDB",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "feature-branch",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_dependabot.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(action)
		require.NoError(t, err)
		expected := result{
			dbBase: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "main",
			},
			dbHead: repoID{
				owner:  "AlekSi",
				repo:   "FerretDB",
				branch: "dependabot/go_modules/tools/github.com/reviewdog/reviewdog-0.14.0",
			},
		}
		assert.Equal(t, expected, actual)
	})
}
