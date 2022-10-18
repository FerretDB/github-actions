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

package main

import (
	"context"
	"errors"
	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"
	"strconv"

	_ "github.com/FerretDB/github-actions/internal"
)

func main() {
	ctx := context.Background()
	action := githubactions.New()
	client := github.NewClient(nil)

	// receive project webhook
	getProjectHook(ctx, action, client)
	// check sig
	// check secret
}

// getProjectHook returns the project hook
func getProjectHook(ctx context.Context, action *githubactions.Action, client *github.Client) (string, error) {
	// TODO: set this
	orgID := action.Getenv("GITHUB_ORG_ID")

	hookID, err := findProjectV2HookID(ctx, action, client)
	if err != nil {
		return "", err
	}
	_, _, err = client.Organizations.GetHook(ctx, orgID, hookID)
	if err != nil {
		return "", err
	}

	return "", errors.New("not implemented")
}

// findProjectV2HookID returns hookID configured on env var.
// It intends to find matching hookID for project_v2_item
func findProjectV2HookID(ctx context.Context, action *githubactions.Action, client *github.Client) (int64, error) {
	hookID, err := strconv.ParseInt(action.Getenv("GITHUB_PROJECT_HOOK_ID"), 10, 64)
	if err != nil {
		return 0, err
	}
	return hookID, nil
}
