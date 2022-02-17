package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	action := githubactions.New()
	client := internal.GitHubClient(ctx, action)

	result, err := detect(ctx, action, client)
	if err != nil {
		internal.DumpEnv(action)
		action.Fatalf("%s", err)
	}

	action.Infof("Detected: %+v.", result)
	action.Noticef("Detected: %s.", result.url)

	action.SetOutput("owner", result.owner)
	action.SetOutput("repo", result.repo)
	if result.branch != "" {
		action.SetOutput("branch", result.branch)
	}
	if result.number != 0 {
		action.SetOutput("number", strconv.Itoa(result.number))
	}
}

// branchID represents a named branch in owner's repo.
type branchID struct {
	owner  string // AlekSi
	repo   string // dance
	branch string // feature-branch
}

type result struct {
	owner  string // AlekSi
	repo   string // dance
	branch string // feature-branch
	number int    // 1
	url    string // https://github.com/AlekSi/dance/tree/feature-branch or https://github.com/AlekSi/dance/pull/1
}

func detect(ctx context.Context, action *githubactions.Action, client *github.Client) (*result, error) {
	event, err := readEvent(action)
	if err != nil {
		return nil, fmt.Errorf("detect: %w", err)
	}

	var base, head branchID

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
			return nil, fmt.Errorf("detect: %w", err)
		}

		base.owner = *event.PullRequest.Base.Repo.Owner.Login
		base.repo = *event.PullRequest.Base.Repo.Name
		base.branch = *event.PullRequest.Base.Ref

		head.owner = *event.PullRequest.Head.Repo.Owner.Login
		head.repo = *event.PullRequest.Head.Repo.Name
		head.branch = *event.PullRequest.Head.Ref

	case *github.PushEvent:
		baseRef := event.GetBaseRef()
		ref := event.GetRef()
		if baseRef != "" || ref != "refs/heads/main" {
			return nil, fmt.Errorf("detect: unhandled push to %q / %q", baseRef, ref)
		}

		base.owner = *event.Repo.Owner.Login
		base.repo = *event.Repo.Name
		base.branch = "main"

		head.owner = *event.Repo.Owner.Login
		head.repo = *event.Repo.Name
		head.branch = "main"

	default:
		return nil, fmt.Errorf("detect: unhandled event type %T", event)
	}

	// figure out the other repo (FerretDB or dance)
	var otherRepo string
	switch base.repo {
	case "dance":
		otherRepo = "FerretDB"
	case "FerretDB":
		otherRepo = "dance"
	default:
		return nil, fmt.Errorf("detect: unhandled repo %q", base.repo)
	}

	pr, err := getPR(ctx, action, client, base.owner, otherRepo, &head)
	if err != nil {
		return nil, fmt.Errorf("detect: %w", err)
	}

	res := &result{
		owner: base.owner,
		repo:  otherRepo,
	}
	if pr == nil {
		res.branch = base.branch
		res.url = fmt.Sprintf("https://github.com/%s/%s/tree/%s", base.owner, otherRepo, base.branch)
	} else {
		res.number = *pr.Number
		res.url = *pr.HTMLURL
	}
	return res, nil
}

// readEvent reads event from GITHUB_EVENT_PATH path.
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
	case "push":
		event = new(github.PushEvent)
	default:
		return nil, fmt.Errorf("unhandled event to unmarshal: %q", eventName)
	}

	if err := json.Unmarshal(b, event); err != nil {
		return nil, err
	}

	return event, nil
}

// getPR returns the first PR in baseOwner/baseRepo from head.owner/head.repo@head.branch.
func getPR(ctx context.Context, action *githubactions.Action, client *github.Client, baseOwner, baseRepo string, head *branchID) (*github.PullRequest, error) {
	action.Infof("Getting PR in %s/%s from %s/%s@%s ...", baseOwner, baseRepo, head.owner, head.repo, head.branch)

	headLabel := head.owner + ":" + head.branch
	opts := &github.PullRequestListOptions{
		State:       "all",
		Head:        headLabel,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		prs, resp, err := client.PullRequests.List(ctx, baseOwner, baseRepo, opts)
		if err != nil {
			return nil, fmt.Errorf("getPR: %w", err)
		}

		for _, pr := range prs {
			action.Debugf("PR: %s.", pr)

			if o := *pr.User.Login; o != head.owner {
				action.Debugf("Unexpected user %q (expected %q).", o, head.owner)
				continue
			}
			if l := *pr.Head.Label; l != headLabel {
				action.Debugf("Unexpected head's label %q (expected %q).", l, headLabel)
				continue
			}
			if o := *pr.Head.User.Login; o != head.owner {
				action.Debugf("Unexpected head's user %q (expected %q).", o, head.owner)
				continue
			}
			if b := *pr.Head.Ref; b != head.branch {
				action.Debugf("Unexpected head's branch %q (expected %q).", b, head.branch)
				continue
			}

			return pr, nil
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	action.Infof("Did not find a matching PR.")
	return nil, nil
}
