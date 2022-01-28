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
	owner, name, tag, err := extract(action, os.Getenv)
	if err != nil {
		// dump environment for debugging
		for _, l := range os.Environ() {
			if strings.HasPrefix(l, "GITHUB_") {
				action.Infof("%s", l)
			}
		}

		action.Fatalf("%s", err)
	}

	action.Noticef("Extracted owner %q, name %q, tag %q.", owner, name, tag)
	action.SetOutput("owner", owner)
	action.SetOutput("name", name)
	action.SetOutput("tag", tag)
}

func extract(_ *githubactions.Action, getEnv func(string) string) (owner, name, tag string, err error) {
	repo := getEnv("GITHUB_REPOSITORY")
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) == 2 {
		owner = parts[0]
		name = parts[1]
	}
	if owner == "" || name == "" {
		err = fmt.Errorf("failed to extract owner or name from %q", repo)
		return
	}

	switch getEnv("GITHUB_EVENT_NAME") {
	case "pull_request", "pull_request_target":
		branch := getEnv("GITHUB_HEAD_REF")
		tag = "dev-" + strings.ToLower(branch) // always add prefix to prevent clashes on "main", "latest", etc
	case "push", "schedule", "workflow_run":
		branch := getEnv("GITHUB_REF_NAME")
		if branch == "main" { // build on pull_request/pull_request_target for other branches
			tag = strings.ToLower(branch)
		}
	}

	if tag == "" {
		err = fmt.Errorf("failed to extract tag")
		return
	}

	return
}
