// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_self.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			number: 1,
			url:    "https://github.com/AlekSi/dance/pull/1",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_fork.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "FerretDB",
			repo:   "dance",
			number: 47,
			url:    "https://github.com/FerretDB/dance/pull/47",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
			url:    "https://github.com/AlekSi/dance/tree/main",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/self", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_target_self.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			number: 1,
			url:    "https://github.com/AlekSi/dance/pull/1",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/fork", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_target_fork.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "FerretDB",
			repo:   "dance",
			number: 47,
			url:    "https://github.com/FerretDB/dance/pull/47",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "pull_request_target_dependabot.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
			url:    "https://github.com/AlekSi/dance/tree/main",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "push.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
			url:    "https://github.com/AlekSi/dance/tree/main",
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("schedule", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_EVENT_PATH": filepath.Join("..", "testdata", "schedule.json"),
			"GITHUB_TOKEN":      "",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := detect(ctx, action, internal.GitHubClient(ctx, action, "GITHUB_TOKEN"))
		require.NoError(t, err)
		expected := &result{
			owner:  "AlekSi",
			repo:   "dance",
			branch: "main",
			url:    "https://github.com/AlekSi/dance/tree/main",
		}
		assert.Equal(t, expected, actual)
	})
}
