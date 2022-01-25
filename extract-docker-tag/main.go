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
	image, tag, err := extractDockerImageTag(action, os.Getenv)
	if err != nil {
		// dump environment for debugging
		for _, l := range os.Environ() {
			if strings.HasPrefix(l, "GITHUB_") {
				action.Infof("%s", l)
			}
		}

		action.Fatalf("%s", err)
	}

	action.Noticef("Extracted image %q, tag %q.", image, tag)
	action.SetOutput("image", image)
	action.SetOutput("tag", tag)
}

func extractDockerImageTag(_ *githubactions.Action, getEnv func(string) string) (image, tag string, err error) {
	repo := getEnv("GITHUB_REPOSITORY")
	image = "ghcr.io/" + strings.ToLower(repo)

	if image == "" {
		err = fmt.Errorf("failed to extract image")
		return
	}

	switch getEnv("GITHUB_EVENT_NAME") {
	case "pull_request":
		branch := getEnv("GITHUB_HEAD_REF")
		tag = "dev-" + strings.ToLower(branch) // always add prefix to prevent clashes on "main", "latest", etc
	case "push", "schedule":
		branch := getEnv("GITHUB_REF_NAME")
		if branch == "main" { // build on pull_request for other branches
			tag = strings.ToLower(branch)
		}
	}

	if tag == "" {
		err = fmt.Errorf("failed to extract tag")
		return
	}

	return
}
