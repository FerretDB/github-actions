package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-githubactions"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GraphQLClient returns GitHub GraphQL client instance with an access token provided from GitHub Actions.
// The token to access API must be provided in the environment variable named `tokenVar`.
func GraphQLClient(ctx context.Context, action *githubactions.Action, tokenVar string) (*githubv4.Client, error) {
	token := action.Getenv(tokenVar)
	if token == "" {
		return nil, fmt.Errorf("env %s is not set", tokenVar)
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
	if err := qlClient.Query(ctx, &rl, nil); err != nil {
		return nil, err
	}
	action.Debugf(
		"Rate limit remaining: %d, reset at: %s",
		rl.RateLimit.Remaining, rl.RateLimit.ResetAt.Format(time.RFC822),
	)

	return qlClient, nil
}
