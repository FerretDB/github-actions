package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	result, err := extract(action, os.Getenv)
	if err != nil {
		// dump environment for debugging
		for _, l := range os.Environ() {
			if strings.HasPrefix(l, "GITHUB_") {
				action.Infof("%s", l)
			}
		}

		action.Fatalf("%s", err)
	}

	action.Noticef("Extracted: %+v.", result)
	action.SetOutput("owner", result.owner)
	action.SetOutput("name", result.name)
	action.SetOutput("tag", result.tag)
	action.SetOutput("ghcr", result.ghcr)
}

type result struct {
	owner string // ferretdb
	name  string // github-actions-dev
	tag   string // pr-add-features
	ghcr  string // ghcr.io/ferretdb/github-actions-dev:pr-add-features
}

func extract(_ *githubactions.Action, getEnv func(string) string) (result result, err error) {
	// set owner and name
	repo := getEnv("GITHUB_REPOSITORY")
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) == 2 {
		result.owner = parts[0]
		result.name = parts[1]
	}
	if result.owner == "" || result.name == "" {
		err = fmt.Errorf("failed to extract owner or name from %q", repo)
		return
	}

	// set tag
	event := getEnv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request", "pull_request_target":
		// always add suffix and prefix to prevent clashes on "main", "latest", etc
		result.name += "-dev"
		branch := getEnv("GITHUB_HEAD_REF")
		result.tag = "pr-" + strings.ToLower(branch)
	case "push", "schedule", "workflow_run":
		branch := getEnv("GITHUB_REF_NAME")
		if branch == "main" { // build on pull_request/pull_request_target for other branches
			result.tag = strings.ToLower(branch)
		}
	}
	if result.tag == "" {
		err = fmt.Errorf("failed to extract tag for event %q", event)
		return
	}

	// set ghcr
	result.ghcr = fmt.Sprintf("ghcr.io/%s/%s:%s", result.owner, result.name, result.tag)

	return
}
