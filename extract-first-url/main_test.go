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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFirstURLFromFile(t *testing.T) {
	t.Run("get_first_url", func(t *testing.T) {

		// action := githubactions.New(githubactions.WithGetenv(getEnv))
		file, err := os.Open(`./test/deploy.txt`)
		if err != nil {
			require.NoError(t, err)
		}
		defer file.Close()

		actual := getFirstURLFromFile(file)
		require.NoError(t, err)

		expected := "https://abc.com"
		assert.Equal(t, expected, actual)
	})
	t.Run("get_first_url_empty", func(t *testing.T) {

		file, err := os.Open(`./test/empty_deploy.txt`)
		if err != nil {
			require.NoError(t, err)
		}
		defer file.Close()
		actual := getFirstURLFromFile(file)
		require.NoError(t, err)

		expected := ""
		assert.Equal(t, expected, actual)
	})
}
