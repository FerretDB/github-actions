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
	"errors"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestCheckTitle(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		expectedErr error
	}{{
		name:        "pull_request/title_without_dot",
		title:       "I'm a title without a dot",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_a_digit",
		title:       "I'm a title without a digit 1",
		expectedErr: nil,
	}, {
		name:        "pull_request/title_with_dot",
		title:       "I'm a title with a dot.",
		expectedErr: errors.New("PR title must end with a latin letter or digit"),
	}, {
		name:        "pull_request/title_with_whitespace",
		title:       "I'm a title with a whitespace ",
		expectedErr: errors.New("PR title must end with a latin letter or digit"),
	}, {
		name:        "pull_request/title_with_backticks",
		title:       "I'm a title with a `backticks`",
		expectedErr: nil,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkTitle(githubactions.New(), tc.title)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestCheckBody(t *testing.T) {
	errNoPunctuation := errors.New("PR body must end with dot or other punctuation mark")

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
