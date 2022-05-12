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

func main() {
	flag.Parse()

	action := githubactions.New()

	internal.DebugEnv(action)

	// check environment variables
	workspace := action.Getenv("GITHUB_WORKSPACE")
	gopath := action.Getenv("GOPATH")
	gocache := action.Getenv("GOCACHE")
	golangciLintCache := action.Getenv("GOLANGCI_LINT_CACHE")
	gomodcache := action.Getenv("GOMODCACHE")
	goproxy := action.Getenv("GOPROXY")

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
	if goproxy == "" {
		action.Fatalf("GOPROXY is not set")
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
		action.Fatalf("GOPROXY must be set to `https://proxy.golang.org`")
	}

	// set parameters for the cache key
	_, week := time.Now().ISOWeek()
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

		// We need to run `go mod download` with `all` argument since 1.18, but that would add extra go.sum entries.
		// Instead, run `go mod tidy` as it is good enough.
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = filepath.Dir(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		start := time.Now()
		action.Infof("Running `%s` in %s ...", strings.Join(cmd.Args, " "), cmd.Dir)
		if err = cmd.Run(); err != nil {
			return err
		}
		action.Infof("Done in %s.", time.Since(start))

		return nil
	})
	if err != nil {
		action.Fatalf("Error walking directory: %s", err)
	}
}
