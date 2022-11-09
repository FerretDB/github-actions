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

package graphql

import (
	"context"
	"encoding/json"

	"github.com/shurcooL/githubv4"
)

// Some documentation URLs might be broken; see https://github.com/github/docs/issues/19568.

// Fields maps field name to field value.
type Fields map[string]string

// PullRequest contains information about a pull request.
type PullRequest struct {
	Title     string
	Body      string
	Author    string
	AuthorBot bool
	Labels    []string

	// ProjectFields maps project title to fields.
	ProjectFields map[string]Fields
}

// https://docs.github.com/en/graphql/reference/interfaces#projectv2fieldcommon
type projectV2FieldCommon struct {
	ID       githubv4.ID
	Name     githubv4.String
	DataType githubv4.ProjectNextFieldType
}

// https://docs.github.com/en/graphql/reference/objects#projectv2iterationfield
type projectV2IterationField struct {
	Configuration *struct {
		Duration githubv4.Int
		StartDay githubv4.Int
	}
}

// https://docs.github.com/en/graphql/reference/objects#projectv2singleselectfield
type projectV2SingleSelectField struct {
	Options []struct {
		ID   githubv4.ID
		Name githubv4.String
	}
}

// https://docs.github.com/en/graphql/reference/interfaces#projectv2itemfieldvaluecommon
type projectV2ItemFieldValueCommon struct {
	ID    githubv4.ID
	Field struct {
		projectV2FieldCommon `graphql:"... on ProjectV2FieldCommon"`
	}
}

// https://docs.github.com/en/graphql/reference/objects#projectv2itemfielditerationvalue
type projectV2ItemFieldIterationValue struct {
	Title     githubv4.String
	Duration  githubv4.Int
	StartDate githubv4.String
}

// https://docs.github.com/en/graphql/reference/objects#projectv2itemfieldsingleselectvalue
type projectV2ItemFieldSingleSelectValue struct {
	OptionId githubv4.String
	Name     githubv4.String
}

type projectField struct {
	Typename                   githubv4.String `graphql:"__typename"`
	projectV2FieldCommon       `graphql:"... on ProjectV2FieldCommon"`
	ProjectV2IterationField    projectV2IterationField    `graphql:"... on ProjectV2IterationField"`
	ProjectV2SingleSelectField projectV2SingleSelectField `graphql:"... on ProjectV2SingleSelectField"`
}

type projectFieldValue struct {
	Typename                            githubv4.String `graphql:"__typename"`
	projectV2ItemFieldValueCommon       `graphql:"... on ProjectV2ItemFieldValueCommon"`
	ProjectV2ItemFieldIterationValue    projectV2ItemFieldIterationValue    `graphql:"... on ProjectV2ItemFieldIterationValue"`
	ProjectV2ItemFieldSingleSelectValue projectV2ItemFieldSingleSelectValue `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
}

// https://docs.github.com/en/graphql/reference/objects#pullrequest
type pullRequest struct {
	Title githubv4.String
	Body  githubv4.String

	// https://docs.github.com/en/graphql/reference/interfaces#actor
	Author struct {
		Typename githubv4.String `graphql:"__typename"`
		Login    githubv4.String
	}

	// https://docs.github.com/en/graphql/reference/interfaces#labelable
	Labels struct {
		Nodes []struct {
			ID   githubv4.ID
			Name githubv4.String
		}
	} `graphql:"labels(first: 20)"`

	// https://docs.github.com/en/graphql/reference/objects#projectv2item
	ProjectItems struct {
		Nodes []struct {
			Typename githubv4.String `graphql:"__typename"`

			ID githubv4.ID

			Project struct {
				ID     githubv4.ID
				Title  githubv4.String
				Fields struct {
					Nodes []projectField
				} `graphql:"fields(first: 20)"`
			}

			FieldValues struct {
				Nodes []projectFieldValue
			} `graphql:"fieldValues(first: 20)"`
		}
	} `graphql:"projectItems(first: 20)"`
}

// GetPullRequest returns information about a pull request by GraphQL node ID.
func (c *Client) GetPullRequest(ctx context.Context, nodeID string) *PullRequest {
	var q struct {
		Node struct {
			PullRequest pullRequest `graphql:"... on PullRequest"`
		} `graphql:"node(id: $nodeID)"`
	}

	variables := map[string]any{
		"nodeID": githubv4.ID(nodeID),
	}

	if err := c.Query(ctx, &q, variables); err != nil {
		c.action.Fatalf("Failed to query pull request: %s.", err)
		return nil
	}

	b, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		c.action.Fatalf("%s", err)
	}

	c.action.Infof("Got:\n%s", b)

	res := &PullRequest{
		Title:  string(q.Node.PullRequest.Title),
		Body:   string(q.Node.PullRequest.Body),
		Author: string(q.Node.PullRequest.Author.Login),
	}

	if q.Node.PullRequest.Author.Typename == "Bot" {
		res.AuthorBot = true
	}

	labelNodes := q.Node.PullRequest.Labels.Nodes
	if len(labelNodes) == 20 {
		c.action.Fatalf("Too many Labels nodes.")
		return nil
	}

	for _, labelNode := range labelNodes {
		res.Labels = append(res.Labels, string(labelNode.Name))
	}

	itemNodes := q.Node.PullRequest.ProjectItems.Nodes
	if len(itemNodes) == 20 {
		c.action.Fatalf("Too many ProjectItems nodes.")
		return nil
	}

	for _, itemNode := range itemNodes {
		fields := make(Fields)

		// checks if IterationField exists and initializes its key in map
		// to handle cases where "Sprint" field exists but not set
		for _, field := range itemNode.Project.Fields.Nodes {
			if field.Typename == "ProjectV2IterationField" {
				fields[string(field.Name)] = ""
			}
		}

		valueNodes := itemNode.FieldValues.Nodes
		if len(valueNodes) == 20 {
			c.action.Fatalf("Too many ProjectItems.FieldValues nodes.")
			return nil
		}

		for _, valueNode := range valueNodes {
			switch valueNode.Typename {
			// Get those values from the pull request itself instead.
			// case "ProjectV2ItemFieldLabelValue":
			// case "ProjectV2ItemFieldMilestoneValue":
			// case "ProjectV2ItemFieldPullRequestValue":
			// case "ProjectV2ItemFieldRepositoryValue":
			// case "ProjectV2ItemFieldReviewerValue":
			// case "ProjectV2ItemFieldUserValue":

			case "ProjectV2ItemFieldDateValue":
			case "ProjectV2ItemFieldIterationValue":
				fields[string(valueNode.Field.Name)] = string(valueNode.ProjectV2ItemFieldIterationValue.Title)
			case "ProjectV2ItemFieldNumberValue":
			case "ProjectV2ItemFieldSingleSelectValue":
				fields[string(valueNode.Field.Name)] = string(valueNode.ProjectV2ItemFieldSingleSelectValue.Name)
			case "ProjectV2ItemFieldTextValue":
			}
		}

		if res.ProjectFields == nil {
			res.ProjectFields = make(map[string]Fields)
		}
		res.ProjectFields[string(itemNode.Project.Title)] = fields
	}

	return res
}
