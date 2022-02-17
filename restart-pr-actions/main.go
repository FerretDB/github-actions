package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

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

// restart restarts actions for PR or branch in action inputs.
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

	workflowRunIDs, err := collectWorkflowRunIDs(ctx, action, client, owner, repo, branch, number)
	if err != nil {
		return fmt.Errorf("restart: %w", err)
	}

	for _, workflowRunID := range workflowRunIDs {
		if err := rerunWorkflow(ctx, action, client, owner, repo, workflowRunID); err != nil {
			return fmt.Errorf("restart: %w", err)
		}
	}

	var allCompleted bool
	for !allCompleted {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}

		allCompleted = true
		for _, workflowRunID := range workflowRunIDs {
			status, conclusion, err := foo(ctx, action, client, owner, repo, workflowRunID)
			if err != nil {
				return fmt.Errorf("restart: %w", err)
			}

			if status != "completed" {
				allCompleted = false
				continue
			}

			if conclusion != "success" {
				action.Errorf("Workflow run %d %s with %s.", workflowRunID, status, conclusion)
				return fmt.Errorf("failed")
			}
		}
	}

	return nil
}

// collectWorkflowRunIDs collects workflow run IDs for a given branch or PR.
func collectWorkflowRunIDs(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo, branch string, number int) ([]int64, error) {
	var headSHA string
	var err error
	if branch != "" {
		headSHA, err = getBranch(ctx, action, client, owner, repo, branch)
	} else {
		headSHA, err = getPR(ctx, action, client, owner, repo, number)
	}
	if err != nil {
		return nil, fmt.Errorf("collectWorkflowRunIDs: %w", err)
	}

	checkRunIDs, err := listCheckRunsForRef(ctx, action, client, owner, repo, headSHA)
	if err != nil {
		return nil, fmt.Errorf("collectWorkflowRunIDs: %w", err)
	}

	// We can't use https://docs.github.com/en/rest/reference/checks#rerequest-a-check-suite API
	// as it is available only for GitHub Apps, and GITHUB_TOKEN (which is a token set by GitHub App,
	// see https://docs.github.com/en/actions/security-guides/automatic-token-authentication) is repo-scoped;
	// we can't use it to access a different repository.
	//
	// Instead, we rely on the fact that Checks API's check run ID matches Actions API's job run ID,
	// and use the latter, that is available for Personal Access Tokens.

	workflowRunIDs := make(map[int64]struct{})
	for _, checkRunID := range checkRunIDs {
		workflowRunID, err := getWorkflowRun(ctx, action, client, owner, repo, checkRunID)
		if err != nil {
			return nil, fmt.Errorf("collectWorkflowRunIDs: %w", err)
		}
		workflowRunIDs[workflowRunID] = struct{}{}
	}

	res := make([]int64, 0, len(workflowRunIDs))
	for workflowRunID := range workflowRunIDs {
		res = append(res, workflowRunID)
	}
	return res, nil
}

// getPR returns PR's head SHA.
func getPR(ctx context.Context, action *githubactions.Action, client *github.Client, baseOwner, baseRepo string, number int) (string, error) {
	pr, _, err := client.PullRequests.Get(ctx, baseOwner, baseRepo, number)
	if err != nil {
		return "", fmt.Errorf("getPR: %w", err)
	}

	action.Debugf("getPR: %s", github.Stringify(pr))

	sha := *pr.Head.SHA
	action.Infof("Got head %s for PR %s.", sha, *pr.HTMLURL)
	return sha, nil
}

// getBranch returns branch's head SHA.
func getBranch(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo, branch string) (string, error) {
	br, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, false)
	if err != nil {
		return "", fmt.Errorf("getBranch: %w", err)
	}

	action.Debugf("getBranch: %s", github.Stringify(br))

	sha := *br.Commit.SHA
	action.Infof("Got head %s for branch %s.", sha, *br.Name)
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

			action.Infof("Found: %q (%d) %s", *checkRun.Name, *checkRun.ID, *checkRun.HTMLURL)
			checkRunIDs = append(checkRunIDs, *checkRun.ID)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return checkRunIDs, nil
}

// getWorkflowRun returns workflow run ID by a job run ID (which is equal to check run ID).
//
// https://docs.github.com/en/rest/reference/actions#get-a-job-for-a-workflow-run
func getWorkflowRun(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, jobRunID int64) (int64, error) {
	job, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobRunID)
	if err != nil {
		return 0, fmt.Errorf("getWorkflowRun: %w", err)
	}

	action.Debugf("getWorkflowRun: %s", github.Stringify(job))

	workflowRunID := *job.RunID
	action.Infof("Found workflow run ID %d for %q (%d)", workflowRunID, *job.Name, *job.ID)
	return workflowRunID, nil
}

// rerunWorkflow re-runs workflow (making a new attempt) by a workflow run ID.
//
// https://docs.github.com/en/rest/reference/actions#re-run-a-workflow
func rerunWorkflow(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, workflowRunID int64) error {
	action.Infof("Restarting workflow run %d ...", workflowRunID)

	if _, err := client.Actions.RerunWorkflowByID(ctx, owner, repo, workflowRunID); err != nil {
		return fmt.Errorf("rerunWorkflow: %w", err)
	}

	return nil
}

// https://docs.github.com/en/rest/reference/actions#get-a-workflow-run
func foo(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, workflowRunID int64) (string, string, error) {
	run, _, err := client.Actions.GetWorkflowRunByID(ctx, owner, repo, workflowRunID)
	if err != nil {
		return "", "", fmt.Errorf("foo: %w", err)
	}

	action.Debugf("foo: %s", github.Stringify(run))

	status := *run.Status
	conclusion := *run.Conclusion
	return status, conclusion, nil
}
