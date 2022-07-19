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
	"github.com/shurcooL/githubv4"
)

// GraphQLFields represent a list of GitHub PNFs (Project Next Field).
type GraphQLFields struct {
	TotalCount githubv4.Int
	Nodes      []GraphQLField
}

// GraphQLField represents a single GitHub PNF (Project Next Field).
type GraphQLField struct {
	ID       githubv4.ID
	Name     githubv4.String
	DataType githubv4.ProjectNextFieldType
	Settings githubv4.String
}

// GraphQLFieldValues represents list of GitHub PNFVs (Project Next Field Value).
type GraphQLFieldValues struct {
	TotalCount githubv4.Int
	Nodes      []GraphQLFieldValue
}

// GraphQLFieldValue represents a single GitHub PNFV (Project Next Field Value).
type GraphQLFieldValue struct {
	ID           githubv4.ID
	Value        githubv4.String
	ProjectField GraphQLField

	// ValueTitle is a special field to display the value in a more human-readable way.
	ValueTitle string `graphql:"value"`
}

// Items represents a list of GitHub PNIs (Project Next Item).
type Items struct {
	TotalCount githubv4.Int
	Nodes      []ProjectV2Item `graphql:"fieldValues(first: $fieldValuesMax)"`
}

type ProjectV2Item struct {
	FieldValues ProjectV2ItemFieldValueConnection
}

type ProjectV2ItemFieldValueConnection struct {
	TotalCount githubv4.Int
	Nodes      []map[string]any
}

type ProjectV2ItemFieldLabelValue struct {
	Labels struct {
		Nodes struct {
			Name string
		}
	} `graphql:"labels(first: $labelsMax)"`
}
type ProjectV2ItemFieldIterationValue struct {
	Title string
}
type ProjectV2ItemFieldMilestoneValue struct {
	Milestone struct {
		Title string
	}
}
type ProjectV2ItemFieldReviewerValue struct {
	Reviewers struct {
		Nodes struct {
			User struct {
				Login string
			}
		}
	} `graphql:"reviewers(first: $reviewersMax)"`
}

type ProjectV2ItemFieldTextValue struct {
	Text string
}
type ProjectV2ItemFieldSingleSelectValue struct {
	Name string
}

type PRItem struct {
	FieldName string
	Value     string
}

// Querier describes a GitHub GraphQL client that can make a query.
type Querier interface {
	// Query executes the given GraphQL query `q` with the given variables `vars` and stores the results in `q`.
	Query(ctx context.Context, q any, vars map[string]any) error
}

// GetPRItems returns the list of PNIs - Project Next Items (cards) associated with the given PR.
func GetPRItems(client Querier, nodeID string) ([]PRItem, error) {
	var q struct {
		Node struct {
			ID           githubv4.String
			Title        githubv4.String
			State        githubv4.String
			ProjectItems Items `graphql:"projectItems(first: $itemsMax)"`
		} `graphql:"node(id: $nodeID)"`
	}

	variables := map[string]any{
		"nodeID":         githubv4.ID(nodeID),
		"itemsMax":       githubv4.Int(20),
		"fieldsMax":      githubv4.Int(20),
		"labelsMax":      githubv4.Int(10),
		"reviewersMax":   githubv4.Int(10),
		"fieldValuesMax": githubv4.Int(10),
	}

	if err := client.Query(context.Background(), &q, variables); err != nil {
		return nil, err
	}

	if q.Node.ProjectItems.TotalCount == 0 {
		return nil, nil
	}

	var result []PRItem

	for _, v := range q.Node.ProjectItems.Nodes[0].FieldValues.Nodes {
		typename, ok := v["__typename"]
		if !ok {
			continue
		}
		switch typename {
		case "ProjectV2ItemFieldIterationValue":
			title, ok := v["title"]
			if !ok {
				continue
			}

			result = append(result, PRItem{
				FieldName: getFieldName(v),
				Value:     title.(string),
			})
		case "ProjectV2ItemFieldMilestoneValue":
			milestone, ok := v["milestone"]
			if !ok {
				continue
			}
			title, ok := milestone.(map[string]any)["title"]
			if !ok {
				continue
			}

			result = append(result, PRItem{
				FieldName: getFieldName(v),
				Value:     title.(string),
			})
		case "ProjectV2ItemFieldSingleSelectValue":
			name, ok := v["name"]
			if !ok {
				continue
			}

			result = append(result, PRItem{
				FieldName: getFieldName(v),
				Value:     name.(string),
			})

		}
	}

	return result, nil
}

func getFieldName(v map[string]any) string {
	field, ok := v["field"]
	if !ok {
		return ""
	}
	fieldName, ok := field.(map[string]any)["name"]
	if !ok {
		return ""
	}

	return fieldName.(string)
}
