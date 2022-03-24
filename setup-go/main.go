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
	if gomodcache == "" {
		action.Fatalf("GOMODCACHE is not set")
	}
	if !strings.HasPrefix(gocache, gopath) {
		action.Fatalf("GOCACHE must be a subdirectory of GOPATH")
	}
	if strings.HasPrefix(gomodcache, gocache) {
		action.Fatalf("GOMODCACHE must not be a subdirectory of GOCACHE")
	}
	if goproxy != "https://proxy.golang.org" {
		action.Fatalf("GOPROXY must be set to `https://proxy.golang.org`")
	}

	// set parameters for the cache key
	_, week := time.Now().ISOWeek()
	action.SetOutput("cache_week", strconv.Itoa(week))
	action.SetOutput("cache_path", gocache)

	// call `go mod download` in directories with `go.mod` file
	err := filepath.Walk(workspace, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		// skip .git, etc
		if info.IsDir() {
			action.Debugf("%s", path)
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		if name != "go.mod" {
			return nil
		}

		cmd := exec.Command("go", "mod", "download")
		cmd.Dir = filepath.Dir(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		action.Infof("Running `go mod download` in %s ...", cmd.Dir)
		err = cmd.Run()
		return err
	})
	if err != nil {
		action.Fatalf("Error walking directory: %s", err)
	}
}
