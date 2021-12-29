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
	tag, err := extractDockerTag(action, os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	action.SetOutput("tag", tag)
}

func extractDockerTag(_ *githubactions.Action, getEnv func(string) string) (string, error) {
	tag := getEnv("GITHUB_HEAD_REF")

	return tag, nil
}
