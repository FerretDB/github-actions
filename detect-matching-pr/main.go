package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DumpEnv(action)

	event, err := readEvent(action)
	if err != nil {
		action.Fatalf("%s", err)
	}
	payload, err := event.ParsePayload()
	if err != nil {
		action.Fatalf("%s", err)
	}
	_ = payload

	result, err := detect(action, os.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	action.Noticef("Detected: %+v.", result)
}

type result struct {
	owner string // AlekSi
}

func detect(_ *githubactions.Action, getEnv func(string) string) (result result, err error) {
	// set owner, get repo
	var repo string
	parts := strings.Split(getEnv("GITHUB_REPOSITORY"), "/")
	if len(parts) == 2 {
		result.owner = parts[0]
		repo = parts[1]
	}
	if result.owner == "" {
		err = fmt.Errorf("failed to detect owner %q", repo)
		return
	}

	event := getEnv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request":
		// branch := getEnv("GITHUB_HEAD_REF")
	default:
		err = fmt.Errorf("unsupported event %q", event)
	}

	return
}

func readEvent(action *githubactions.Action) (*github.Event, error) {
	name := os.Getenv("GITHUB_EVENT_PATH")
	if name == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	action.Debugf("Read event from %s:\n%s", name, string(b))

	var event github.Event
	if err = json.Unmarshal(b, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func getClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc), nil
}
