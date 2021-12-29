package main

import (
	"flag"
	"log"
	"os"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	if _, err := extractDockerTag(action, os.Getenv); err != nil {
		log.Fatal(err)
	}
}

func extractDockerTag(action *githubactions.Action, getEnv func(string) string) (string, error) {
	tag := getEnv("GITHUB_HEAD_REF")

	action.Infof("Extracted tag %q.", tag)
	action.SetOutput("tag", tag)

	return tag, nil
}
