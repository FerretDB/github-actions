package main

import (
	"context"
	"flag"
	"fmt"
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

	action.Infof("Getting check suites for %s/%s@%s ...", owner, repo, headSHA)

	opts := &github.ListCheckSuiteOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		suites, resp, err := client.Checks.ListCheckSuitesForRef(ctx, owner, repo, headSHA, opts)
		if err != nil {
			return fmt.Errorf("restartPRActions: %w", err)
		}

		for _, suite := range suites.CheckSuites {
			action.Debugf("Check suite: %s.", suite)

			action.Infof("Restarting check suite %s ...", *suite.URL)
			if _, err = client.Checks.ReRequestCheckSuite(ctx, owner, repo, *suite.ID); err != nil {
				return fmt.Errorf("restartPRActions: %w", err)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// action.Infof("Waiting for %s check suites to complete ...", *pr.HTMLURL)

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
			suites, resp, err := client.Checks.ListCheckSuitesForRef(ctx, owner, repo, headSHA, opts)
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

	return nil
}
