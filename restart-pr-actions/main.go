package main

import (
	"context"
	"flag"
	"fmt"
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

	if err := restart(ctx, action, client); err != nil {
		internal.DumpEnv(action)
		action.Fatalf("%s", err)
	}
}

// restart restarts actions for PR in action inputs.
func restart(ctx context.Context, action *githubactions.Action, client *github.Client) error {
	owner := action.GetInput("owner")
	if owner == "" {
		return fmt.Errorf("restart: owner is required")
	}
	repo := action.GetInput("repo")
	if repo == "" {
		return fmt.Errorf("restart: repo is required")
	}
	branch := action.GetInput("branch")
	numberS := action.GetInput("number")
	if (branch == "") == (numberS == "") {
		return fmt.Errorf("restart: exactly one of branch and number should be set")
	}

	number, err := strconv.Atoi(numberS)
	if err != nil {
		return fmt.Errorf("restart: %w", err)
	}

	var headSHA string
	if branch != "" {
		headSHA, err = getBranch(ctx, action, client, owner, repo, branch)
	} else {
		headSHA, err = getPR(ctx, action, client, owner, repo, number)
	}
	if err != nil {
		return fmt.Errorf("restart: %w", err)
	}

	checkRunIDs, err := listCheckRunsForRef(ctx, action, client, owner, repo, headSHA)
	if err != nil {
		return fmt.Errorf("restart: %w", err)
	}

	// We can't use https://docs.github.com/en/rest/reference/checks#rerequest-a-check-suite
	// as it is available only for GitHub Apps.
	// Instead, we rely on the fact that check run ID matches Actions job ID.

	runIDs := make(map[int64]struct{})
	for _, checkRunID := range checkRunIDs {
		runID, err := foo(ctx, action, client, owner, repo, checkRunID)
		if err != nil {
			return fmt.Errorf("restart: %w", err)
		}
		runIDs[runID] = struct{}{}
	}

	for runID := range runIDs {
		if err := bar(ctx, action, client, owner, repo, runID); err != nil {
			return fmt.Errorf("restart: %w", err)
		}
	}

	// TODO wait

	// action.Infof("Waiting for %s check suites to complete ...", *pr.HTMLURL)

	/*
		var allCompleted bool
		for !allCompleted {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}

			allCompleted = true
			opts.Page = 0

			for {
				suites, resp, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, headSHA, opts)
				if err != nil {
					return fmt.Errorf("restart: %w", err)
				}

				for _, suite := range suites.CheckSuites {
					action.Debugf("Check suite: %s.", suite)

					if *suite.Status != "completed" {
						allCompleted = false
						continue
					}

					if *suite.Conclusion != "success" {
						action.Infof("Check suite %d %s with %q.", *suite.ID, *suite.Status, *suite.Conclusion)
						return fmt.Errorf("some checks failed")
						// return fmt.Errorf("Some %s checks failed.", *pr.HTMLURL)
					}
				}

				if resp.NextPage == 0 {
					break
				}
				opts.Page = resp.NextPage
			}
		}
	*/

	return nil
}

// getPR returns PR's head SHA.
func getPR(ctx context.Context, action *githubactions.Action, client *github.Client, baseOwner, baseRepo string, number int) (string, error) {
	pr, _, err := client.PullRequests.Get(ctx, baseOwner, baseRepo, number)
	if err != nil {
		return "", fmt.Errorf("getPR: %w", err)
	}

	sha := *pr.Head.SHA
	action.Infof("Got %q for %s.", sha, *pr.HTMLURL)
	return sha, nil
}

// getBranch returns branch's head SHA.
func getBranch(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo, branch string) (string, error) {
	br, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, false)
	if err != nil {
		return "", fmt.Errorf("getBranch: %w", err)
	}

	sha := *br.Commit.SHA
	action.Infof("Got %q for %s.", sha, *br.Name)
	return sha, nil
}

// listCheckRunsForRef returns GitHub Actions check run IDs for given PR (owner/repo@headSHA).
//
// https://docs.github.com/en/rest/reference/checks#list-check-runs-for-a-git-reference
func listCheckRunsForRef(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo, headSHA string) ([]int64, error) {
	action.Infof("Getting GitHub Actions check run IDs for %s/%s@%s ...", owner, repo, headSHA)

	var checkRunIDs []int64
	opts := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		checkRuns, resp, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, headSHA, opts)
		if err != nil {
			return nil, fmt.Errorf("listCheckRunsForRef: %w", err)
		}

		for _, checkRun := range checkRuns.CheckRuns {
			action.Debugf("Check run: %s.", github.Stringify(checkRun))

			if *checkRun.App.Slug != "github-actions" {
				continue
			}

			action.Infof("Found: %d %s %s", *checkRun.ID, *checkRun.Name, *checkRun.HTMLURL)
			checkRunIDs = append(checkRunIDs, *checkRun.ID)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return checkRunIDs, nil
}

// https://docs.github.com/en/rest/reference/actions#get-a-job-for-a-workflow-run
func foo(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, jobID int64) (int64, error) {
	action.Infof("foo: jobID = %d", jobID)

	job, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobID)
	if err != nil {
		return 0, fmt.Errorf("foo: %w", err)
	}

	action.Debugf("foo: workflow job: %s", github.Stringify(job))

	action.Infof("foo: runID = %d", *job.RunID)

	return *job.RunID, nil
}

// https://docs.github.com/en/rest/reference/actions#re-run-a-workflow
func bar(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, runID int64) error {
	action.Infof("bar: runID = %d", runID)

	if _, err := client.Actions.RerunWorkflowByID(ctx, owner, repo, runID); err != nil {
		return fmt.Errorf("bar: %w", err)
	}

	return nil
}
