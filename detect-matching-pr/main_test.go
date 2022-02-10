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
		result, err := detect(action)
		require.NoError(t, err)
		assert.Equal(t, "AlekSi", result.owner)
	})

	t.Run("pull_request_target/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("testdata", "pull_request_target_self.json"),
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		result, err := detect(action)
		require.NoError(t, err)
		assert.Equal(t, "AlekSi", result.owner)
	})
}
