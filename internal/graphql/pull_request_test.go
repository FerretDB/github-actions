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
	"time"

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
			Labels:    []string{"deps"},
		}
		actual := c.GetPullRequest(ctx, "PR_kwDOGfwnTc48nVkp")
		actual.Body, _, _ = strings.Cut(actual.Body, "\n")
		assert.Equal(t, expected, actual)
	})

	// https://github.com/FerretDB/github-actions/pull/85
	t.Run("ProjectV2", func(t *testing.T) {
		t.Parallel()

		expected := &PullRequest{
			Title:  "Migrate to `ProjectV2`",
			Body:   "Test body.",
			Author: "AlekSi",
			Labels: []string{"code/chore", "trust"},
			ProjectFields: map[string]Fields{
				"Test project": {
					"Size":   "S",
					"Status": "Done",
				},
				"Another test project": {
					"Size":   "🐋 X-Large",
					"Sprint": "",
					"Status": "🔖 Ready",
				},
			},
		}
		actual := c.GetPullRequest(ctx, "PR_kwDOGfwnTc48u60R")
		assert.Equal(t, expected, actual)
	})
}

func TestIsCurrentIteration(t *testing.T) {
	startDate := func(n int) string {
		return time.Now().Add(time.Duration(n*24) * time.Hour).Format("2006-01-02")
	}
	cases := []struct {
		name      string
		startDate string
		want      bool
	}{{
		name:      "ok",
		startDate: startDate(-7),
		want:      true,
	}, {
		name:      "equal_start_date",
		startDate: startDate(0),
		want:      true,
	}, {
		name:      "equal_end_date",
		startDate: startDate(-14),
		want:      false, // difference in milliseconds causing this to be false
	}, {
		name:      "after_end_date",
		startDate: startDate(-15),
		want:      false,
	}, {
		name:      "before_start_date",
		startDate: startDate(15),
		want:      false,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isCurrentIteration(tc.startDate, 14)) // fix 14 duration for testing purpose only
		})
	}
}
