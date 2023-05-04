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
	"errors"
	"fmt"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/github-actions/internal"
	"github.com/FerretDB/github-actions/internal/graphql"
)

func TestRunPRChecks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	action := githubactions.New()
	c := &checker{
		action:  action,
		client:  internal.GitHubClient(ctx, action, "GITHUB_TOKEN"),
		gClient: graphql.NewClient(ctx, action, "CONFORM_TOKEN"),
	}

	// To get node ID from PR:
	// curl https://api.github.com/repos/FerretDB/github-actions/pulls/83 | jq '.node_id'

	cases := []struct {
		name              string
		user              string
		nodeID            string
		expectedRes       []checkResult
		expectedCommunity bool
	}{{
		name:              "Dependabot",
		user:              "dependabot",
		nodeID:            "PR_kwDOGfwnTc48nVkp", // https://github.com/FerretDB/github-actions/pull/83
		expectedCommunity: true,
	}, {
		name:   "OneProjectAllFieldsUnset",
		user:   "AlekSi",
		nodeID: "PR_kwDOGfwnTc48tuFy", // https://github.com/FerretDB/github-actions/pull/84
		expectedRes: []checkResult{
			{check: "Labels"},
			{check: "Size"},
			{check: "Sprint", err: fmt.Errorf(`PR should have "Sprint" field set.`)},
			{check: "Title"},
			{check: "Body"},
			{check: "Auto-merge"},
		},
	}, {
		name:   "TwoProjectsMix",
		user:   "AlekSi",
		nodeID: "PR_kwDOGfwnTc48u60R", // https://github.com/FerretDB/github-actions/pull/85
		expectedRes: []checkResult{
			{check: "Labels"},
			{
				check: "Size",
				err:   fmt.Errorf(`PR should have "Size" field unset, got "🐋 X-Large" for project "Another test project".`),
			},
			{check: "Sprint"},
			{check: "Title"},
			{check: "Body"},
			{check: "Auto-merge"},
		},
	}, {
		name:   "Community",
		user:   "ronaudinho",
		nodeID: "PR_kwDOGfwnTc5BT7Ej", // https://github.com/FerretDB/github-actions/pull/109
		expectedRes: []checkResult{
			{check: "Labels"},
			{check: "Size"},
			{check: "Sprint"},
			{check: "Title"},
			{check: "Body"},
			{check: "Auto-merge"},
		},
		expectedCommunity: true,
	}, {
		name:   "AutoMerge",
		user:   "AlekSi",
		nodeID: "PR_kwDOGfwnTc5DpH8i", // https://github.com/FerretDB/github-actions/pull/120
		expectedRes: []checkResult{
			{
				check: "Labels",
				err:   fmt.Errorf(`That PR should not be merged yet.`),
			},
			{check: "Size"},
			{
				check: "Sprint",
				err:   fmt.Errorf(`PR should have "Sprint" field set.`),
			},
			{
				check: "Title",
				err:   fmt.Errorf(`PR title must end with a latin letter or digit.`),
			},

			{check: "Body"},
			{check: "Auto-merge"},
		},
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, community := c.runChecks(ctx, "FerretDB", tc.user, tc.nodeID)
			assert.Equal(t, tc.expectedRes, res)
			assert.Equal(t, tc.expectedCommunity, community)
		})
	}
}

func TestCheckTitle(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		expectedErr error
	}{{
		name:        "pull_request/title_without_dot",
		title:       "Test the title without a dot",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_a_digit",
		title:       "Test the title without a digit 1",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_dot",
		title:       "I'm a title with a dot.",
		expectedErr: errors.New("PR title must end with a latin letter or digit."),
	}, {
		name:        "pull_request/title_with_whitespace",
		title:       "I'm a title with a whitespace ",
		expectedErr: errors.New("PR title must end with a latin letter or digit."),
	}, {
		name:        "pull_request/title_with_backticks",
		title:       "Test the title I'm a title with a `backticks`",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_without_uppercase",
		title:       "test the title that does not start with an uppercase`",
		expectedErr: errors.New("PR title must start with an uppercase letter."),
	}, {
		name:        "pull_request/title_with_imperative_verb",
		title:       "Fix `$` path errors for sort",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_imperative_verb",
		title:       "Document `not ready` issues label",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_imperative_verb",
		title:       "Bump deps",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_invalid_imperative_verb",
		title:       "Please do not merge this PR",
		expectedErr: nil, // `prose` fails to detect this as a verb
	}, {
		name:        "pull_request/title_with_invalid_imperative_verb",
		title:       "A title without an imperative verb at the beginning",
		expectedErr: nil, // `prose` fails to detect this as a verb
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkTitle(githubactions.New(), tc.title)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestCheckBody(t *testing.T) {
	errNoPunctuation := errors.New("PR body must end with dot or other punctuation mark.")

	cases := []struct {
		name        string
		body        string
		expectedErr error
	}{{
		name:        "pull_request/empty_body",
		body:        "",
		expectedErr: nil,
	}, {
		name:        "pull_request/whitespace_body",
		body:        "\n",
		expectedErr: errNoPunctuation,
	}, {
		name:        "pull_request/body_with_dot",
		body:        "I'm a body with a dot.",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_!",
		body:        "I'm a body with a punctuation mark!\r\n",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_with_?",
		body:        "Am I a body with a punctuation mark?",
		expectedErr: nil,
	}, {
		name:        "pull_request/body_without_dot",
		body:        "I'm a body without a dot\n",
		expectedErr: errNoPunctuation,
	}, {
		name:        "pull_request/body_too_shot",
		body:        "!\r\n",
		expectedErr: errNoPunctuation,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkBody(githubactions.New(), tc.body)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
