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
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/github-actions/internal/testutil"
)

func TestExtractFerretDB(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:pr-extract-docker-tag",
				"ghcr.io/ferretdb/all-in-one:pr-extract-docker-tag",
				"quay.io/ferretdb/all-in-one:pr-extract-docker-tag",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:pr-extract-docker-tag",
				"ghcr.io/ferretdb/ferretdb-dev:pr-extract-docker-tag",
				"quay.io/ferretdb/ferretdb-dev:pr-extract-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:pr-extract-docker-tag",
				"ghcr.io/ferretdb/all-in-one:pr-extract-docker-tag",
				"quay.io/ferretdb/all-in-one:pr-extract-docker-tag",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:pr-extract-docker-tag",
				"ghcr.io/ferretdb/ferretdb-dev:pr-extract-docker-tag",
				"quay.io/ferretdb/ferretdb-dev:pr-extract-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
			"GITHUB_REF_NAME":   "58/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:pr-mongo-go-driver-29d768e",
				"ghcr.io/ferretdb/all-in-one:pr-mongo-go-driver-29d768e",
				"quay.io/ferretdb/all-in-one:pr-mongo-go-driver-29d768e",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:pr-mongo-go-driver-29d768e",
				"ghcr.io/ferretdb/ferretdb-dev:pr-mongo-go-driver-29d768e",
				"quay.io/ferretdb/ferretdb-dev:pr-mongo-go-driver-29d768e",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/main", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:main",
				"ghcr.io/ferretdb/all-in-one:main",
				"quay.io/ferretdb/all-in-one:main",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:main",
				"ghcr.io/ferretdb/ferretdb-dev:main",
				"quay.io/ferretdb/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/release", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "releases/1.2",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:releases-1.2",
				"ghcr.io/ferretdb/all-in-one:releases-1.2",
				"quay.io/ferretdb/all-in-one:releases-1.2",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:releases-1.2",
				"ghcr.io/ferretdb/ferretdb-dev:releases-1.2",
				"quay.io/ferretdb/ferretdb-dev:releases-1.2",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/beta", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v0.1.0-beta",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:0.1.0-beta",
				"ghcr.io/ferretdb/all-in-one:0.1.0-beta",
				"quay.io/ferretdb/all-in-one:0.1.0-beta",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:0.1.0-beta",
				"ghcr.io/ferretdb/ferretdb-dev:0.1.0-beta",
				"quay.io/ferretdb/ferretdb-dev:0.1.0-beta",
			},
			productionImages: []string{
				"ferretdb/ferretdb:0.1.0-beta",
				"ghcr.io/ferretdb/ferretdb:0.1.0-beta",
				"quay.io/ferretdb/ferretdb:0.1.0-beta",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/release", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v0.1.0",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:0.1.0",
				"ferretdb/all-in-one:latest",
				"ghcr.io/ferretdb/all-in-one:0.1.0",
				"ghcr.io/ferretdb/all-in-one:latest",
				"quay.io/ferretdb/all-in-one:0.1.0",
				"quay.io/ferretdb/all-in-one:latest",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:0.1.0",
				"ferretdb/ferretdb-dev:latest",
				"ghcr.io/ferretdb/ferretdb-dev:0.1.0",
				"ghcr.io/ferretdb/ferretdb-dev:latest",
				"quay.io/ferretdb/ferretdb-dev:0.1.0",
				"quay.io/ferretdb/ferretdb-dev:latest",
			},
			productionImages: []string{
				"ferretdb/ferretdb:0.1.0",
				"ferretdb/ferretdb:latest",
				"ghcr.io/ferretdb/ferretdb:0.1.0",
				"ghcr.io/ferretdb/ferretdb:latest",
				"quay.io/ferretdb/ferretdb:0.1.0",
				"quay.io/ferretdb/ferretdb:latest",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/wrong", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "0.1.0", // no leading v
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		_, err := extract(action)
		require.Error(t, err)
	})

	t.Run("schedule", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:main",
				"ghcr.io/ferretdb/all-in-one:main",
				"quay.io/ferretdb/all-in-one:main",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:main",
				"ghcr.io/ferretdb/ferretdb-dev:main",
				"quay.io/ferretdb/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:main",
				"ghcr.io/ferretdb/all-in-one:main",
				"quay.io/ferretdb/all-in-one:main",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:main",
				"ghcr.io/ferretdb/ferretdb-dev:main",
				"quay.io/ferretdb/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})
}

func TestExtractOther(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "extract-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/some-repo",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			developmentImages: []string{
				"ghcr.io/ferretdb/some-repo-dev:pr-extract-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/main", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/some-repo",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			developmentImages: []string{
				"ghcr.io/ferretdb/some-repo-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/release", func(t *testing.T) {
		getEnv := testutil.GetEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v0.1.0",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/some-repo",
		})

		action := githubactions.New(githubactions.WithGetenv(getEnv))
		actual, err := extract(action)
		require.NoError(t, err)

		expected := &result{
			developmentImages: []string{
				"ghcr.io/ferretdb/some-repo-dev:0.1.0",
				"ghcr.io/ferretdb/some-repo-dev:latest",
			},
			productionImages: []string{
				"ghcr.io/ferretdb/some-repo:0.1.0",
				"ghcr.io/ferretdb/some-repo:latest",
			},
		}
		assert.Equal(t, expected, actual)
	})
}

//nolint:lll // it is more readable this way
func TestImageURL(t *testing.T) {
	// expected URLs should work
	assert.Equal(t, "https://ghcr.io/ferretdb/all-in-one:pr-extract-docker-tag", imageURL("ghcr.io/ferretdb/all-in-one:pr-extract-docker-tag"))
	assert.Equal(t, "https://quay.io/ferretdb/all-in-one:pr-extract-docker-tag", imageURL("quay.io/ferretdb/all-in-one:pr-extract-docker-tag"))
	assert.Equal(t, "https://hub.docker.com/r/ferretdb/all-in-one/tags", imageURL("ferretdb/all-in-one:pr-extract-docker-tag"))
}
