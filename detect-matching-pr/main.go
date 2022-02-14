package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	result, err := detect(context.Background(), action)
	if err != nil {
		internal.DumpEnv(action)
		action.Fatalf("%s", err)
	}

	action.Noticef("Detected: %+v.", result)
	action.SetOutput("db_base_owner", result.dbBase.owner)
	action.SetOutput("db_base_repo", result.dbBase.repo)
	action.SetOutput("db_base_branch", result.dbBase.branch)
	action.SetOutput("db_head_owner", result.dbHead.owner)
	action.SetOutput("db_head_repo", result.dbHead.repo)
	action.SetOutput("db_head_branch", result.dbHead.branch)
}

type repoID struct {
	owner  string // FerretDB
	repo   string // dance
	branch string // main
}

type result struct {
	dbBase    repoID // FerretDB/FerretDB@main
	dbHead    repoID // AlekSi/FerretDB@feature-branch
	danceBase repoID // FerretDB/dance@main
	danceHead repoID // AlekSi/dance@feature-branch
}

func detect(ctx context.Context, action *githubactions.Action) (res *result, err error) {
	res = new(result)

	var event interface{}
	if event, err = readEvent(action); err != nil {
		return
	}

	var base, head repoID

	// extract information from event
	switch event := event.(type) {
	case *github.PullRequestEvent:
		// check that author sends PR from own repo
		switch {
		case *event.Sender.Login == "dependabot[bot]":
			// nothing, that's a special case
		case *event.Sender.Login != *event.PullRequest.User.Login:
			err = fmt.Errorf(
				"event.Sender.Login %q != event.PullRequest.User.Login %q",
				*event.Sender.Login, *event.PullRequest.User.Login,
			)
		case *event.Sender.Login != *event.PullRequest.Head.User.Login:
			err = fmt.Errorf(
				"event.Sender.Login %q != event.PullRequest.Head.User.Login %q",
				*event.Sender.Login, *event.PullRequest.Head.User.Login,
			)
		case *event.Sender.Login != *event.PullRequest.Head.Repo.Owner.Login:
			err = fmt.Errorf(
				"event.Sender.Login %q != event.PullRequest.Head.Repo.Owner.Login %q",
				*event.Sender.Login, *event.PullRequest.Head.Repo.Owner.Login,
			)
		}
		if err != nil {
			return
		}

		base.owner = *event.PullRequest.Base.Repo.Owner.Login
		base.repo = *event.PullRequest.Base.Repo.Name
		base.branch = *event.PullRequest.Base.Ref

		head.owner = *event.PullRequest.Head.Repo.Owner.Login
		head.repo = *event.PullRequest.Head.Repo.Name
		head.branch = *event.PullRequest.Head.Ref

	default:
		err = fmt.Errorf("unhandled event type %T", event)
	}

	// figure out the repo (FerretDB or dance)
	ferretdbRepo := regexp.MustCompile(`(?i)ferretdb`)
	danceRepo := regexp.MustCompile(`(?i)dance`)
	var otherRepo *regexp.Regexp
	switch r := strings.ToLower(base.repo); {
	case strings.Contains(r, "dance"): // check first to allow renaming to something like "FerretDB-dance"
		res.dbBase = base
		res.dbHead = head
		otherRepo = danceRepo
	case strings.Contains(r, "ferretdb"):
		res.dbBase = base
		res.dbHead = head
		otherRepo = ferretdbRepo
	default:
		err = fmt.Errorf("unhandled repo %q", r)
		return
	}

	client := getClient(ctx, action)
	getRepo(ctx, client, base.owner, otherRepo)

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

	// Debug level requires `ACTIONS_RUNNER_DEBUG` secret to be set to `true`:
	// https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/enabling-debug-logging
	// Note that `pull_request` events from forks do not have access to secrets,
	// so that line will not be logged in that case.
	action.Debugf("Read event from %s:\n%s", eventPath, string(b))

	eventName := action.Getenv("GITHUB_EVENT_NAME")
	if eventName == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_NAME is not set")
	}

	var event interface{}
	switch eventName {
	case "pull_request", "pull_request_target":
		event = new(github.PullRequestEvent)
	default:
		return nil, fmt.Errorf("unhandled event to unmarshal: %q", eventName)
	}

	if err := json.Unmarshal(b, event); err != nil {
		return nil, err
	}

	return event, nil
}

// getClient returns GitHub API client with token from enviroment, if present.
func getClient(ctx context.Context, action *githubactions.Action) *github.Client {
	token := action.Getenv("GITHUB_TOKEN")
	if token == "" {
		action.Debugf("GITHUB_TOKEN is not set")
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

// getRepo returns owner's GitHub repo that matches the given name regexp.
func getRepo(ctx context.Context, client *github.Client, owner string, name *regexp.Regexp) (*github.Repository, error) {
	opts := &github.RepositoryListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.List(ctx, owner, opts)
		if err != nil {
			return nil, fmt.Errorf("getRepo: %w", err)
		}

		for _, repo := range repos {
			if name.MatchString(pointer.GetString(repo.Name)) {
				return repo, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil, fmt.Errorf("getRepo: failed to find a matching repo")
}
