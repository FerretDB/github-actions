package gh

import (
	"context"

	"github.com/sethvargo/go-githubactions"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GraphQLClient returns GitHub GraphQL client instance with an access token provided from GitHub Actions.
func GraphQLClient(ctx context.Context, action *githubactions.Action) *githubv4.Client {
	token := action.Getenv("GITHUB_TOKEN")
	if token == "" {
		action.Debugf("GITHUB_TOKEN is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)
	return githubv4.NewClient(httpClient)
}
