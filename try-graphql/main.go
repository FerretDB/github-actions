package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_GRAPHQL_TEST_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// query some project information
	{
		var q struct {
			Organization struct {
				Project struct {
					Title  githubv4.String
					Fields struct {
						Nodes []struct {
							ID       githubv4.String
							Name     githubv4.String
							DataType githubv4.String
							Settings githubv4.String
						}

						TotalCount githubv4.Int
						PageInfo   struct {
							EndCursor   githubv4.String
							HasNextPage githubv4.Boolean
						}
					} `graphql:"fields(first: $fieldsMax)"`
				} `graphql:"projectNext(number: $projectID)"`
			} `graphql:"organization(login: $organization)"`

			RateLimit struct {
				Cost      githubv4.Int
				Limit     githubv4.Int
				Remaining githubv4.Int
				ResetAt   githubv4.DateTime
			}
		}

		variables := map[string]interface{}{
			"organization": githubv4.String("FerretDB"),
			"projectID":    githubv4.Int(3),
			"fieldsMax":    githubv4.Int(20),
		}

		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			return err
		}
		printJSON(q)

		// returns settings for the project field by its name
		getSettings := func(name githubv4.String) githubv4.String {
			for _, node := range q.Organization.Project.Fields.Nodes {
				if node.Name != name {
					continue
				}

				return node.Settings
			}

			return ""
		}

		// get sprint settings from the field
		settings := getSettings("Sprint")

		var sprintSettings struct {
			Configuration struct {
				Duration   int `json:"duration"`
				StartDay   int `json:"start_day"`
				Iterations []struct {
					ID        string `json:"id"`
					Title     string `json:"title"`
					Duration  int    `json:"duration"`
					StartDate string `json:"start_date"`
					TitleHTML string `json:"title_html"`
				} `json:"iterations"`
				CompletedIterations []struct {
					ID        string `json:"id"`
					Title     string `json:"title"`
					Duration  int    `json:"duration"`
					StartDate string `json:"start_date"`
					TitleHTML string `json:"title_html"`
				} `json:"completed_iterations"`
			} `json:"configuration"`
		}

		err = json.Unmarshal([]byte(settings), &sprintSettings)
		if err != nil {
			return err
		}
		printJSON(sprintSettings)
	}

	return nil
}

// printJSON prints v as JSON encoded with indent to stdout. It panics on any error.
func printJSON(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}
