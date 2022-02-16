package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	action := githubactions.New()
	client := internal.GitHubClient(ctx, action)

	if err := restartPRActions(ctx, action, client); err != nil {
		internal.DumpEnv(action)
		action.Fatalf("%s", err)
	}
}

// restartPRActions restarts actions for PR in action inputs.
func restartPRActions(ctx context.Context, action *githubactions.Action, client *github.Client) error {
	var owner, repo, number, headSHA string
	if owner = action.GetInput("owner"); owner == "" {
		return fmt.Errorf("restartPRActions: owner is required")
	}
	if repo = action.GetInput("repo"); repo == "" {
		return fmt.Errorf("restartPRActions: repo is required")
	}
	if number = action.GetInput("number"); number == "" {
		return fmt.Errorf("restartPRActions: number is required")
	}
	if headSHA = action.GetInput("head_sha"); headSHA == "" {
		return fmt.Errorf("restartPRActions: head_sha is required")
	}

	checkRunIDs, err := listCheckRunsForRef(ctx, action, client, owner, repo, headSHA)
	if err != nil {
		return fmt.Errorf("restartPRActions: %w", err)
	}

	// We can't use https://docs.github.com/en/rest/reference/checks#rerequest-a-check-suite
	// as it is available only for GitHub Apps.
	// Instead, we rely on the fact that check run ID matches Actions job ID.

	var failed bool
	for _, checkRunID := range checkRunIDs {
		if err = restartJob(ctx, action, client, owner, repo, checkRunID); err != nil {
			action.Errorf("restartPRActions: %s", err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("restartPRActions: some steps failed")
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
					return fmt.Errorf("restartPRActions: %w", err)
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

// restartJob gets workflow job by ID, extract its run ID and restart it.
//
// https://docs.github.com/en/rest/reference/actions#get-a-job-for-a-workflow-run
// https://docs.github.com/en/rest/reference/actions#re-run-a-workflow
func restartJob(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, jobID int64) error {
	action.Infof("restartJob: jobID = %d", jobID)

	job, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobID)
	if err != nil {
		return fmt.Errorf("restartRun: %w", err)
	}

	action.Debugf("restartJob: workflow job: %s", github.Stringify(job))

	action.Infof("restartJob: jobID = %d", *job.RunID)

	if _, err = client.Actions.RerunWorkflowByID(ctx, owner, repo, *job.RunID); err != nil {
		return fmt.Errorf("restartRun: %w", err)
	}

	return nil
}
