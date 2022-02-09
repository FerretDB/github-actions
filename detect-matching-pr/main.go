package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	_ = event

	result, err := detect(action)
	if err != nil {
		action.Fatalf("%s", err)
	}

	action.Noticef("Detected: %+v.", result)
}

type result struct {
	owner string // AlekSi
}

func detect(action *githubactions.Action) (result result, err error) {
	// set owner, get repo
	var repo string
	parts := strings.Split(action.Getenv("GITHUB_REPOSITORY"), "/")
	if len(parts) == 2 {
		result.owner = parts[0]
		repo = parts[1]
	}
	if result.owner == "" {
		err = fmt.Errorf("failed to detect owner %q", repo)
		return
	}

	event := action.Getenv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request":
		// branch := getEnv("GITHUB_HEAD_REF")
	default:
		err = fmt.Errorf("unsupported event %q", event)
	}

	return
}

func readEvent(action *githubactions.Action) (interface{}, error) {
	eventPath := action.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	b, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return nil, err
	}

	action.Debugf("Read event from %s:\n%s", eventPath, string(b))

	eventName := action.Getenv("GITHUB_EVENT_NAME")
	if eventName == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_NAME is not set")
	}
	return unmarshalEvent(eventName, b)
}

func unmarshalEvent(eventName string, b []byte) (interface{}, error) {
	var event interface{}
	switch eventName {
	case "pull_request":
		event = new(github.PullRequestEvent)
	default:
		return nil, fmt.Errorf("unhandled event %q", eventName)
	}

	if err := json.Unmarshal(b, event); err != nil {
		return nil, err
	}

	return event, nil
}

func getClient(action *githubactions.Action) (*github.Client, error) {
	token := action.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc), nil
}
