package main

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/AlekSi/pointer"
	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()
	internal.DebugEnv(action)

	summaries := runChecks(action)

	for _, summary := range summaries {
		action.AddStepSummary(fmt.Sprintf("%s %b %s", summary.Name, summary.Ok, &summary.Details))
	}

	action.Fatalf("The PR does not conform to the rules")
}

type Summary struct {
	Name    string
	Ok      bool
	Details error
}

// runChecks runs all the checks included into the PR conformance rules.
// It returns the list of errors that occurred during the checks.
func runChecks(action *githubactions.Action) []Summary {
	pr, summaries := getPR(action)
	if len(summaries) > 0 {
		return summaries
	}

	// PRs from dependabot are perfect
	if pr.author == "dependabot[bot]" {
		return nil
	}

	summaries = pr.checkTitle()

	summaries = append(summaries, pr.checkBody()...)

	return summaries
}

// getPR returns PR's information and returns
// * pull request details if no errors
// * a summaries list whether and which check passed successfully or not.
func getPR(action *githubactions.Action) (*pullRequest, []Summary) {
	event, err := internal.ReadEvent(action)
	if err != nil {
		return nil, []Summary{{Name: "Read event", Details: err}}
	}

	var pr pullRequest
	switch event := event.(type) {
	case *github.PullRequestEvent:
		pr.author = *event.PullRequest.User.Login
		pr.title = *event.PullRequest.Title
		pr.body = pointer.Get(event.PullRequest.Body)
	default:
		return nil, []Summary{{
			Name:    "Event type",
			Details: fmt.Errorf("unhandled event type %T (only PR-related events are handled)", event),
		}}
	}
	return &pr, []Summary{}
}

// pullRequest contains information about PR that is interesting for us.
type pullRequest struct {
	author string
	title  string
	body   string
}

// checkTitle checks if PR's title does not end with dot and returns a varying result list for summary.
func (pr *pullRequest) checkTitle() []Summary {
	var results []Summary
	match, err := regexp.MatchString(`[a-zA-Z0-9]$`, pr.title)
	if err != nil {
		results = append(results, Summary{Name: "Title regex parsing", Details: err})
	}

	titleMatches := Summary{Name: "PR title must end with a latin letter or digit"}
	if match {
		titleMatches.Ok = true
	}
	results = append(results, titleMatches)

	return results
}

// checkBody checks if PR's body (description) ends with a punctuation mark.
func (pr *pullRequest) checkBody() []Summary {
	// it is allowed to have an empty body
	if len(pr.body) == 0 {
		return []Summary{}
	}

	match, err := regexp.MatchString(`.+[.!?]$`, pr.body)
	if err != nil {
		return []Summary{{Name: "Body regex parsing", Details: err}}
	}

	bodyCheck := Summary{Name: "PR body must end with dot or other punctuation mark"}
	if match {
		bodyCheck.Ok = true
	}
	return []Summary{bodyCheck}
}
