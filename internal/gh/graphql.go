package gh

import (
	"context"
	"fmt"

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

	// check that the client is able to make queries
	err := qlClient.Query(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	return qlClient, nil
}
