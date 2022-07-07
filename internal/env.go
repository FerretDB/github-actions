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

package internal

import (
	"os"
	"sort"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// DebugEnv logs all environment variables that start with `GITHUB_` or `INPUT_`
// in debug level.
func DebugEnv(action *githubactions.Action) {
	res := make([]string, 0, 30)
	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") || strings.HasPrefix(l, "INPUT_") {
			res = append(res, l)
		}
	}

	sort.Strings(res)

	action.Debugf("Dumping environment variables:")
	for _, l := range res {
		action.Debugf("\t%s", l)
	}
}
