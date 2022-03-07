package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()

	internal.DebugEnv(action)

	result, err := extract(action)
	if err != nil {
		action.Fatalf("%s", err)
	}

	action.Infof("Extracted: %+v.", result)
	action.Noticef("Extracted: https://%s", result.ghcr)

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

func extract(action *githubactions.Action) (result result, err error) {
	// set owner and name
	repo := action.Getenv("GITHUB_REPOSITORY")
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) == 2 {
		result.owner = parts[0]
		result.name = parts[1]
	}
	if result.owner == "" || result.name == "" {
		err = fmt.Errorf("failed to extract owner or name from %q", repo)
		return
	}

	// change name for dance repo
	switch result.name {
	case "dance":
		result.name = "ferretdb"
	case "ferretdb":
		// nothing
	default:
		err = fmt.Errorf("unhandled repo %q", repo)
		return
	}

	// set tag
	event := action.Getenv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request", "pull_request_target":
		// always add suffix and prefix to prevent clashes on "main", "latest", etc
		result.name += "-dev"
		branch := action.Getenv("GITHUB_HEAD_REF")
		parts = strings.Split(strings.ToLower(branch), "/") // for branches like "dependabot/submodules/XXX"
		result.tag = "pr-" + parts[len(parts)-1]
	case "push", "schedule", "workflow_run":
		branch := action.Getenv("GITHUB_REF_NAME")
		if branch == "main" { // build on pull_request/pull_request_target for other branches
			result.name += "-dev"
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
