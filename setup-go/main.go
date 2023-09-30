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
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

// tidyDir runs `go mod tidy -v` in the specified directory.
func tidyDir(action *githubactions.Action, dir string) {
	cmd := exec.Command("go", "mod", "tidy", "-v")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()

	action.Infof("Running `%s` in %s ...", strings.Join(cmd.Args, " "), cmd.Dir)

	if err := cmd.Run(); err != nil {
		action.Fatalf("%s", err)
	}

	action.Infof("Done in %s.", time.Since(start))
}

// checkEnv verifies that environment variables are set correctly.
//
//nolint:wsl // to group things better
func checkEnv(action *githubactions.Action) (workspace, gocache string) {
	workspace = action.Getenv("GITHUB_WORKSPACE")
	gopath := action.Getenv("GOPATH")
	gocache = action.Getenv("GOCACHE")
	golangciLintCache := action.Getenv("GOLANGCI_LINT_CACHE")
	gomodcache := action.Getenv("GOMODCACHE")
	goproxy := action.Getenv("GOPROXY")
	gotoolchain := action.Getenv("GOTOOLCHAIN")

	if workspace == "" {
		action.Fatalf("GITHUB_WORKSPACE is not set")
	}
	if gopath == "" {
		action.Fatalf("GOPATH is not set")
	}

	if gocache == "" {
		action.Fatalf("GOCACHE is not set")
	}
	if golangciLintCache == "" {
		action.Fatalf("GOLANGCI_LINT_CACHE is not set")
	}
	if gomodcache == "" {
		action.Fatalf("GOMODCACHE is not set")
	}

	if !strings.HasPrefix(gocache, gopath) {
		action.Fatalf("GOCACHE must be a subdirectory of GOPATH")
	}
	if !strings.HasPrefix(golangciLintCache, gocache) {
		action.Fatalf("GOLANGCI_LINT_CACHE must be a subdirectory of GOCACHE")
	}
	if strings.HasPrefix(gomodcache, gocache) {
		action.Fatalf("GOMODCACHE must not be a subdirectory of GOCACHE")
	}

	if goproxy != "https://proxy.golang.org" {
		action.Fatalf("GOPROXY must be explicitly set to `https://proxy.golang.org` (without `direct`)")
	}
	if gotoolchain != "local" {
		action.Fatalf("GOTOOLCHAIN must be explicitly set to `local` (without `auto`)")
	}

	return
}

func main() {
	flag.Parse()

	action := githubactions.New()

	internal.DebugEnv(action)
	workspace, gocache := checkEnv(action)

	// set parameters for the cache key
	_, week := time.Now().UTC().ISOWeek() // starts on Monday
	action.SetOutput("cache_week", "w"+strconv.Itoa(week))
	action.SetOutput("cache_path", gocache)

	// download modules in directories with `go.mod` file
	err := filepath.Walk(workspace, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		// skip .git, vendor, etc
		if info.IsDir() {
			action.Debugf("%s", path)
			if strings.HasPrefix(name, ".") || name == "vendor" {
				return filepath.SkipDir
			}
		}

		if name != "go.mod" {
			return nil
		}

		tidyDir(action, filepath.Dir(path))

		return nil
	})
	if err != nil {
		action.Fatalf("Error walking directory: %s", err)
	}
}
