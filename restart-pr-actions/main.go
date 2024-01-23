// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	action := githubactions.New()
	client := internal.GitHubClient(action, "GITHUB_TOKEN")

	internal.DebugEnv(action)

	if err := restart(ctx, action, client); err != nil {
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
	if numberS != "" && err != nil {
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

	action.Infof("Waiting for workflows to finish ...")

	var allCompleted bool
	for !allCompleted {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}

		allCompleted = true
		for _, workflowRunID := range workflowRunIDs {
			run, err := getWorkflowRun(ctx, action, client, owner, repo, workflowRunID)
			if err != nil {
				return fmt.Errorf("restart: %w", err)
			}

			status := *run.Status
			if status != "completed" {
				allCompleted = false
				continue
			}

			conclusion := *run.Conclusion
			if conclusion != "success" {
				action.Errorf("Workflow run %s %s with %s.", *run.HTMLURL, status, conclusion)
				return fmt.Errorf(conclusion)
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
		workflowRunID, err := getWorkflowRunIDByJobRunID(ctx, action, client, owner, repo, checkRunID)
		if err != nil {
			return nil, fmt.Errorf("collectWorkflowRunIDs: %w", err)
		}

		workflowRunIDs[workflowRunID] = struct{}{}
	}

	res := make([]int64, 0, len(workflowRunIDs))
	for workflowRunID := range workflowRunIDs {
		run, err := getWorkflowRun(ctx, action, client, owner, repo, workflowRunID)
		if err != nil {
			return nil, fmt.Errorf("collectWorkflowRunIDs: %w", err)
		}

		// there could be multiple workflow runs for a single ref for different events
		if *run.Event == "schedule" {
			action.Infof("Skipping workflow run ID %d (%q).", workflowRunID, *run.Event)
			continue
		}

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
	action.Noticef("PR: %s", *pr.HTMLURL)
	return sha, nil
}

// getBranch returns branch's head SHA.
func getBranch(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo, branch string) (string, error) {
	br, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, 5)
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

// getWorkflowRunIDByJobRunID returns workflow run ID by a job run ID (which is equal to check run ID).
//
// https://docs.github.com/en/rest/reference/actions#get-a-job-for-a-workflow-run
func getWorkflowRunIDByJobRunID(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, jobRunID int64) (int64, error) {
	job, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobRunID)
	if err != nil {
		return 0, fmt.Errorf("getWorkflowRunIDByJobRunID: %w", err)
	}

	action.Debugf("getWorkflowRunIDByJobRunID: %s", github.Stringify(job))

	workflowRunID := *job.RunID
	action.Infof("Found workflow run ID %d for %q (%d)", workflowRunID, *job.Name, *job.ID)
	return workflowRunID, nil
}

// rerunWorkflow stops and re-runs workflow (making a new attempt) by a workflow run ID.
//
// https://docs.github.com/en/rest/reference/actions#re-run-a-workflow
func rerunWorkflow(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, workflowRunID int64) error {
	run, err := getWorkflowRun(ctx, action, client, owner, repo, workflowRunID)
	if err != nil {
		return fmt.Errorf("rerunWorkflow (get): %w", err)
	}

	name := *run.Name
	url := *run.HTMLURL
	action.Noticef("Workflow run: %q %s", name, url)

	if _, err = client.Actions.CancelWorkflowRunByID(ctx, owner, repo, workflowRunID); err != nil {
		switch githubErr := err.(type) {
		case *github.ErrorResponse:
			// that's the best we can do - er.Errors, er.Block are nil
			if githubErr.Response.StatusCode == 409 {
				switch githubErr.Message {
				case "Cannot cancel a workflow run that is completed.":
					fallthrough
				case "Cannot cancel a workflow re-run that has not yet queued.":
					action.Warningf("rerunWorkflow (cancel): %s", err)
					err = nil
				}
			}

		case *github.AcceptedError:
			action.Warningf("rerunWorkflow (cancel): %s", err)
			err = nil
		}
	}
	if err != nil {
		return fmt.Errorf("rerunWorkflow (cancel): %[1]T %[1]w", err)
	}

	if _, err := client.Actions.RerunWorkflowByID(ctx, owner, repo, workflowRunID); err != nil {
		return fmt.Errorf("rerunWorkflow (rerun): %w", err)
	}

	return nil
}

// getWorkflowRun returns workflow run by workflow run ID.
//
// https://docs.github.com/en/rest/reference/actions#get-a-workflow-run
func getWorkflowRun(ctx context.Context, action *githubactions.Action, client *github.Client, owner, repo string, workflowRunID int64) (*github.WorkflowRun, error) {
	run, _, err := client.Actions.GetWorkflowRunByID(ctx, owner, repo, workflowRunID)
	if err != nil {
		return nil, fmt.Errorf("getWorkflowRun: %w", err)
	}

	action.Debugf("getWorkflowRun: %s", github.Stringify(run))
	return run, nil
}
