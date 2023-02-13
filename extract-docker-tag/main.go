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
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()

	internal.DebugEnv(action)

	result, err := extract(action)
	if err != nil {
		action.Fatalf("%s", err)
	}

	action.Infof("Extracted: %+v.", result)
	action.Noticef("Extracted: https://%s", result.ghcr)

	action.SetOutput("owner", result.owner)
	action.SetOutput("name", result.name)
	action.SetOutput("tag", result.tag)
	action.SetOutput("ghcr", result.ghcr)
	action.SetOutput("ghcr_latest", result.ghcrLatest)
}

type result struct {
	owner      string // ferretdb
	name       string // github-actions-dev
	tags        []string // {"pr-add-features", "0.0.1"}
	ghcr       string // ghcr.io/ferretdb/github-actions-dev:pr-add-features or ghcr.io/ferretdb/github-actions-dev:0.0.1
	ghcrImages []string // {"ghcr.io/ferretdb/github-actions-dev:pr-add-features"} or {"ghcr.io/ferretdb/github-actions-dev:0.0.1", "ghcr.io/ferretdb/github-actions-dev:latest"}
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with leading `v`
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func extract(action *githubactions.Action) (*result, error) {
	result := new(result)

	// set owner and name
	repo := action.Getenv("GITHUB_REPOSITORY")
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) == 2 {
		result.owner = parts[0]
		result.name = parts[1]
	}
	if result.owner == "" || result.name == "" {
		return nil, fmt.Errorf("failed to extract owner or name from %q", repo)
	}

	// set tag, add "-dev" to name if needed
	event := action.Getenv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request", "pull_request_target":
		// always add tag prefix and name suffix to prevent clashes on "main", "latest", etc
		branch := action.Getenv("GITHUB_HEAD_REF")
		parts = strings.Split(strings.ToLower(branch), "/") // for branches like "dependabot/submodules/XXX"
		result.tag = "pr-" + parts[len(parts)-1]
		result.name += "-dev"

	case "push", "schedule", "workflow_run":
		refType := action.Getenv("GITHUB_REF_TYPE")
		refName := action.Getenv("GITHUB_REF_NAME")

		switch refType {
		case "branch":
			// build on pull_request/pull_request_target for other branches
			if refName != "main" {
				return nil, fmt.Errorf("unhandled branch %q", refName)
			}
			result.tag = refName
			result.name += "-dev"

		case "tag":
			match := semVerTag.FindStringSubmatch(refName)
			if match == nil || len(match) != semVerTag.NumSubexp()+1 {
				return nil, fmt.Errorf("unexpected git tag %q", refName)
			}
			result.name += "-dev" // TODO remove for https://github.com/FerretDB/FerretDB/issues/70
			major := match[semVerTag.SubexpIndex("major")]
			minor := match[semVerTag.SubexpIndex("minor")]
			patch := match[semVerTag.SubexpIndex("patch")]
			prerelease := match[semVerTag.SubexpIndex("prerelease")]
			result.tag = major + "." + minor + "." + patch
			if prerelease != "" {
				result.tag += "-" + prerelease
			}

			// add latest for pushed tags
			result.ghcrLatest = fmt.Sprintf("ghcr.io/%s/%s:latest", result.owner, result.name)
		default:
			return nil, fmt.Errorf("unhandled ref type %q", refType)
		}

	default:
		return nil, fmt.Errorf("unhandled event type %q", event)
	}

	if result.tag == "" {
		return nil, fmt.Errorf("failed to extract tag for event %q", event)
	}

	result.ghcr = fmt.Sprintf("ghcr.io/%s/%s:%s", result.owner, result.name, result.tag)
	return result, nil
}
