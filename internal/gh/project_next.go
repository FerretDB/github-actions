package gh

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
	DataType githubv4.String
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
}

// GraphQLItems represents a list of GitHub PNIs (Project Next Item).
type GraphQLItems struct {
	TotalCount githubv4.Int
	Nodes      []GraphQLItem
}

// GraphQLItem represents a single GitHub PNI (Project Next Item).
type GraphQLItem struct {
	ID          githubv4.ID
	FieldValues GraphQLFieldValues `graphql:"fieldValues(first: $fieldsMax)"`
}

// Querier describes a GitHub GraphQL client that can make a query.
type Querier interface {
	Query(ctx context.Context, q interface{}, vars map[string]interface{}) error
}

// GetPRItems returns the list of PNIs - Project Next Items (cards) associated with the given PR.
func GetPRItems(client Querier, nodeID string) ([]GraphQLItem, error) {
	var q struct {
		Node struct {
			PullRequest struct {
				ID                githubv4.String
				ProjectsNextItems GraphQLItems `graphql:"projectNextItems(first: $itemsMax)"`
			} `graphql:"... on PullRequest"`
		} `graphql:"node(id: $nodeID)"`
	}

	variables := map[string]interface{}{
		"nodeID":    githubv4.ID(nodeID),
		"itemsMax":  githubv4.Int(20),
		"fieldsMax": githubv4.Int(20),
	}

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	if q.Node.PullRequest.ProjectsNextItems.TotalCount == 0 {
		return []GraphQLItem{}, nil
	}

	return q.Node.PullRequest.ProjectsNextItems.Nodes, nil
}
