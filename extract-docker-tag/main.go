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
	"strconv"
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

	for _, image := range result.images {
		action.Noticef("https://%s", image)
	}
	action.Noticef("dev: %v", result.dev)

	action.SetOutput("images", strings.Join(result.images, ","))
	action.SetOutput("dev", strconv.FormatBool(result.dev))
}

type result struct {
	images []string
	dev    bool
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func extract(action *githubactions.Action) (*result, error) {
	// extract owner and name to support GitHub forks
	repo := action.Getenv("GITHUB_REPOSITORY")
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to extract owner or name from %q", repo)
	}
	owner := parts[0]
	name := parts[1]

	// extract tags for various events
	event := action.Getenv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request", "pull_request_target":
		// for branches like "dependabot/submodules/XXX"
		branch := strings.ToLower(action.Getenv("GITHUB_HEAD_REF"))
		parts = strings.Split(branch, "/")
		branch = parts[len(parts)-1]

		return &result{
			images: []string{
				fmt.Sprintf("ghcr.io/%s/%s-dev:pr-%s", owner, name, branch),
			},
			dev: true,
		}, nil

	case "push", "schedule", "workflow_run":
		refType := strings.ToLower(action.Getenv("GITHUB_REF_TYPE"))
		refName := strings.ToLower(action.Getenv("GITHUB_REF_NAME"))

		switch refType {
		case "branch":
			// build on pull_request/pull_request_target for other branches
			if refName != "main" {
				return nil, fmt.Errorf("unhandled branch %q", refName)
			}

			return &result{
				images: []string{
					fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, refName),
				},
				dev: true,
			}, nil

		case "tag":
			// extract version from git tag
			match := semVerTag.FindStringSubmatch(refName)
			if match == nil || len(match) != semVerTag.NumSubexp()+1 {
				return nil, fmt.Errorf("unexpected git tag %q", refName)
			}
			major := match[semVerTag.SubexpIndex("major")]
			minor := match[semVerTag.SubexpIndex("minor")]
			patch := match[semVerTag.SubexpIndex("patch")]
			prerelease := match[semVerTag.SubexpIndex("prerelease")]

			version := major + "." + minor + "." + patch
			if prerelease != "" {
				version += "-" + prerelease
			}

			if prerelease != "" {
				return &result{
					images: []string{
						fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, version),
						fmt.Sprintf("ghcr.io/%s/%s-dev:latest", owner, name),
					},
					dev: true,
				}, nil
			}

			res := &result{
				images: []string{
					fmt.Sprintf("ghcr.io/%s/%s:%s", owner, name, version),
					fmt.Sprintf("ghcr.io/%s/%s:latest", owner, name),
				},
				dev: false,
			}

			// https://hub.docker.com/r/ferretdb/ferretdb - no forks, no branches or PRs, only release tags
			if owner == "ferretdb" && name == "ferretdb" {
				res.images = append(
					res.images,
					fmt.Sprintf("ferretdb/ferretdb:%s", version),
					"ferretdb/ferretdb:latest",
				)
			}

			return res, nil

		default:
			return nil, fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		return nil, fmt.Errorf("unhandled event type %q", event)
	}
}
