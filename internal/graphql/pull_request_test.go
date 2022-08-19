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

package graphql

import (
	"context"
	"strings"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestPullRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := NewClient(ctx, githubactions.New(), "CONFORM_TOKEN")

	// To get node ID from PR:
	// curl https://api.github.com/repos/FerretDB/github-actions/pulls/83 | jq '.node_id'

	// https://github.com/FerretDB/github-actions/pull/83
	t.Run("Dependabot", func(t *testing.T) {
		t.Parallel()

		expected := &PullRequest{
			Title:     "Bump github.com/go-task/task/v3 from 3.14.0 to 3.14.1 in /tools",
			Body:      "Bumps [github.com/go-task/task/v3](https://github.com/go-task/task) from 3.14.0 to 3.14.1.",
			Author:    "dependabot",
			AuthorBot: true,
		}
		actual := c.GetPullRequest(ctx, "PR_kwDOGfwnTc48nVkp")
		actual.Body, _, _ = strings.Cut(actual.Body, "\n")
		assert.Equal(t, expected, actual)
	})

	// https://github.com/FerretDB/github-actions/pull/85
	t.Run("ProjectV2", func(t *testing.T) {
		t.Parallel()

		expected := &PullRequest{
			Title:         "Migrate to `ProjectV2`",
			Body:          "Test body.",
			Author:        "AlekSi",
			ProjectFields: map[string]Fields{"Test project": {"Status": "Done"}},
		}
		actual := c.GetPullRequest(ctx, "PR_kwDOGfwnTc48u60R")
		assert.Equal(t, expected, actual)
	})
}
