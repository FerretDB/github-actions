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
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DebugEnv(action)

	path := filepath.Join(action.Getenv("GITHUB_WORKSPACE"), "deploy.txt")
	if u := extractURL(action, path); u != "" {
		action.SetOutput("extracted_url", u)
	}
}

func extractURL(action *githubactions.Action, path string) string {
	f, err := os.Open(path)
	if err != nil {
		action.Fatalf("%s", err)
	}
	defer f.Close()

	re := regexp.MustCompile(`(https?://[^\s]+)`)
	s := bufio.NewScanner(f)

	for s.Scan() {
		if u := re.FindString(s.Text()); u != "" {
			return u
		}
	}

	return ""
}
