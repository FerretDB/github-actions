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
	if _, err := extractDockerTag(action, os.Getenv); err != nil {
		// dump environment for debugging
		for _, l := range os.Environ() {
			if strings.HasPrefix(l, "GITHUB_") {
				action.Infof("%s", l)
			}
		}

		action.Fatalf("%s", err)
	}
}

func extractDockerTag(action *githubactions.Action, getEnv func(string) string) (string, error) {
	var tag string
	switch getEnv("GITHUB_EVENT_NAME") {
	case "pull_request":
		branch := getEnv("GITHUB_HEAD_REF")
		tag = "dev-" + branch // always add prefix to prevent clashes on "main", "latest", etc
	case "push":
		branch := getEnv("GITHUB_REF_NAME")
		if branch == "main" { // build on pull_request for other branches
			tag = branch
		}
	}

	if tag == "" {
		return "", fmt.Errorf("failed to extract tag")
	}

	action.Noticef("Extracted tag %q.", tag)
	action.SetOutput("tag", tag)

	return tag, nil
}
