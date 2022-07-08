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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIterationTitleByID(t *testing.T) {
	activeMany := `{"configuration":{"duration":14,"start_day":1, "iterations":
[{"id":"5f065395","title":"Sprint 1","duration":14,"start_date":"2022-06-20","title_html":"Sprint 1"},
{"id":"a0f5e3ae","title":"Sprint 2","duration":14,"start_date":"2022-07-04","title_html":"Sprint 2"},
{"id":"fd9fd832","title":"Sprint 3","duration":14,"start_date":"2022-07-18","title_html":"Sprint 3"}],
"completed_iterations":[]}}`
	activeOne := `{"configuration":{"duration":14,"start_day":1,"iterations":
[{"id":"5f065395","title":"Sprint 1","duration":14,"start_date":"2022-06-20","title_html":"Sprint 1"}],
"completed_iterations":
[{"id":"a0f5e3ae","title":"Sprint 2","duration":14,"start_date":"2022-07-04","title_html":"Sprint 2"},
{"id":"fd9fd832","title":"Sprint 3","duration":14,"start_date":"2022-07-18","title_html":"Sprint 3"}]}}`

	tc := []struct {
		name          string
		id            string
		settings      string
		expectedTitle string
		expectedErr   bool
	}{
		{
			name:          "active_many_first",
			id:            "5f065395",
			settings:      activeMany,
			expectedTitle: "Sprint 1",
		},
		{
			name:          "active_many_second",
			id:            "a0f5e3ae",
			settings:      activeMany,
			expectedTitle: "Sprint 2",
		},
		{
			name:          "active_one_first",
			id:            "5f065395",
			settings:      activeOne,
			expectedTitle: "Sprint 1",
		},
		{
			name:          "active_one_second",
			id:            "a0f5e3ae",
			settings:      activeOne,
			expectedTitle: "Sprint 2",
		},
		{
			name:          "invalid_id",
			id:            "invalid",
			settings:      activeMany,
			expectedTitle: "",
			expectedErr:   true,
		},
		{
			name:          "invalid_json",
			id:            "5f065395",
			settings:      "{hello}",
			expectedTitle: "",
			expectedErr:   true,
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			title, err := GetIterationTitleByID(tc.id, tc.settings)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedTitle, title)
		})
	}
}

func TestGetSingleSelectTitleByID(t *testing.T) {
	settings := `{"options":
[{"id":"f75ad846","name":"Todo","name_html":"Todo"},
{"id":"47fc9ee4","name":"In Progress","name_html":"In Progress"},
{"id":"98236657","name":"Done","name_html":"Done"}]}`

	tc := []struct {
		name          string
		id            string
		settings      string
		expectedTitle string
		expectedErr   bool
	}{
		{
			name:          "valid_id_first",
			id:            "f75ad846",
			settings:      settings,
			expectedTitle: "Todo",
		},
		{
			name:          "valid_id_second",
			id:            "47fc9ee4",
			settings:      settings,
			expectedTitle: "In Progress",
		},
		{
			name:        "invalid_id",
			id:          "invalid",
			settings:    settings,
			expectedErr: true,
		},
		{
			name:        "invalid_json",
			id:          "f75ad846",
			settings:    `{hello}`,
			expectedErr: true,
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			title, err := GetSingleSelectTitleByID(tc.id, tc.settings)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedTitle, title)
		})
	}
}
