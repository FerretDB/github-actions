package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	result, err := extract(action, os.Getenv)
	if err != nil {
		internal.DumpEnv(action)
		action.Fatalf("%s", err)
	}

	action.Noticef("Extracted: %+v.", result)
}

type result struct{}

func extract(_ *githubactions.Action, getEnv func(string) string) (result result, err error) {
	event := getEnv("GITHUB_EVENT_NAME")
	switch event {
	default:
		err = fmt.Errorf("unsupported event %q", event)
	}

	return
}
