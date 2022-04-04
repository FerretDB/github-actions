package main

import (
	"flag"
	"fmt"
	"regexp"
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
	if result.version != "" {
		action.SetOutput("version", result.version)
	}
	action.SetOutput("ghcr", result.ghcr)
}

type result struct {
	owner   string // ferretdb
	name    string // github-actions-dev
	tag     string // pr-add-features
	version string // semver.org
	ghcr    string // ghcr.io/ferretdb/github-actions-dev:pr-add-features
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
		result.tag, result.version, err = getTag(action)
		if err != nil {
			return
		}
		if branch == "main" || // build on pull_request/pull_request_target for other branches
			action.Getenv("GITHUB_REF_TYPE") == "tag" {
			result.name += "-dev"
		}
	}

	if result.tag == "" {
		err = fmt.Errorf("failed to extract tag for event %q", event)
		return
	}
	result.ghcr = fmt.Sprintf("ghcr.io/%s/%s:%s", result.owner, result.name, result.tag)

	return
}

func getTag(action *githubactions.Action) (tag, version string, err error) {
	refName := action.Getenv("GITHUB_REF_NAME")
	tag = strings.ToLower(refName)

	if action.Getenv("GITHUB_REF_TYPE") != "tag" {
		return
	}

	var semVerRe *regexp.Regexp
	semVerRe, err = regexp.Compile(`(\d+)\.(\d+)\.(\d+)-?([a-zA-Z-\d\.]*)\+?([a-zA-Z-\d\.]*)`)
	if err != nil {
		err = fmt.Errorf("regexp.Compile %w", err)
		return
	}
	version = string(semVerRe.Find([]byte(refName)))
	if version == "" {
		err = fmt.Errorf("tag is not in semver format %q", refName)
		return
	}
	tag = version
	return
}
