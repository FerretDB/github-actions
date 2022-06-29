package graphql

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubQuerier implements graphql.Querier interface for testing purposes.
// It stores a path to the file representing the GraphQL query result.
type stubQuerier string

// Query implements graphql.Querier interface.
func (sq stubQuerier) Query(ctx context.Context, v any, vars map[string]any) error {
	file, err := os.Open(string(sq))
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(&v)
}

func TestGetPRItems(t *testing.T) {
	tc := []struct {
		name        string
		path        string
		expectedLen int
	}{{
		name:        "with_items",
		path:        "pull_request_with_project_items.json",
		expectedLen: 1,
	}, {
		name:        "without_items",
		path:        "pull_request_without_project_items.json",
		expectedLen: 0,
	}}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			client := stubQuerier(filepath.Join("..", "..", "testdata", "graphql", tc.path))

			items, err := GetPRItems(client, "test")
			require.NoError(t, err)
			require.Len(t, items, tc.expectedLen)

			// check that human-readable value is set
			for _, item := range items {
				for _, value := range item.FieldValues.Nodes {
					assert.NotEmpty(t, value.ValueTitle)
				}
			}
		})
	}
}
