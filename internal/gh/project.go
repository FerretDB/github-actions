package gh

import (
	"context"
	"encoding/json"

	"github.com/shurcooL/githubv4"
)

// Project describes a GitHub project (beta).
type Project struct {
	Title          string
	CurrentSprints []Sprint
}

// Sprint describes a GitHub project (beta) sprint (iteration).
type Sprint struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Duration  int    `json:"duration"`
	StartDate string `json:"start_date"`
	TitleHTML string `json:"title_html"`
}

// GraphQLFields describes GraphQL representation of GitHub Project's fields.
type GraphQLFields struct {
	TotalCount githubv4.Int
	Nodes      []struct {
		ID       githubv4.String
		Name     githubv4.String
		DataType githubv4.String
		Settings githubv4.String
	}
}

// Querier describes a GitHub GraphQL client that can make a query.
type Querier interface {
	Query(ctx context.Context, q interface{}, vars map[string]interface{}) error
}

// GetPRProjects returns the list of projects (beta) that the given PR is part of.
func GetPRProjects(client Querier, nodeID string) ([]*Project, error) {
	var q struct {
		Node struct {
			PullRequest struct {
				ID           githubv4.String
				ProjectsNext struct {
					TotalCount githubv4.Int
					Nodes      []struct {
						Title  githubv4.String
						Fields GraphQLFields `graphql:"fields(first: $fieldsMax)"`
					}
				} `graphql:"projectsNext(first: $projectsMax)"`
			} `graphql:"... on PullRequest"`
		} `graphql:"node(id: $nodeID)"`
	}

	variables := map[string]interface{}{
		"nodeID":      githubv4.String(nodeID),
		"projectsMax": githubv4.Int(20),
		"fieldsMax":   githubv4.Int(100),
	}

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	if q.Node.PullRequest.ProjectsNext.TotalCount == 0 {
		return []*Project{}, nil
	}

	projects := make([]*Project, q.Node.PullRequest.ProjectsNext.TotalCount)
	for i, project := range q.Node.PullRequest.ProjectsNext.Nodes {
		sprintSettingsString := getFieldSettings(&project.Fields, "Sprint")

		var sprintSettings struct {
			Configuration struct {
				Duration            int      `json:"duration"`
				StartDay            int      `json:"start_day"`
				Iterations          []Sprint `json:"iterations"`
				CompletedIterations []Sprint `json:"completed_iterations"`
			} `json:"configuration"`
		}
		err = json.Unmarshal([]byte(sprintSettingsString), &sprintSettings)

		if err != nil {
			return nil, err
		}

		projects[i] = &Project{
			Title:          string(project.Title),
			CurrentSprints: sprintSettings.Configuration.Iterations,
		}
	}

	return projects, nil
}

// getSettings returns settings for the project field by its name.
func getFieldSettings(fields *GraphQLFields, name githubv4.String) githubv4.String {
	for _, node := range fields.Nodes {
		if node.Name != name {
			continue
		}

		return node.Settings
	}

	return ""
}
