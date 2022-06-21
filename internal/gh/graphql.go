package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-githubactions"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GraphQLClient returns GitHub GraphQL client instance with an access token provided from GitHub Actions.
func GraphQLClient(ctx context.Context, action *githubactions.Action) (*githubv4.Client, error) {
	token := action.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)
	qlClient := githubv4.NewClient(httpClient)

	// check that the client is able to make queries,
	// for that we call a simple rate limit query
	var rl struct {
		RateLimit struct {
			Cost      githubv4.Int
			Limit     githubv4.Int
			Remaining githubv4.Int
			ResetAt   githubv4.DateTime
		}
	}
	err := qlClient.Query(ctx, &rl, nil)
	if err != nil {
		return nil, err
	}
	action.Infof(
		"Rate limit remaining: %d, reset at: %s",
		rl.RateLimit.Remaining, rl.RateLimit.ResetAt.Format(time.RFC822),
	)

	return qlClient, nil
}
