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

	for _, image := range result.releaseImages {
		action.Noticef("Release: https://%s", image)
	}

	for _, image := range result.developmentImages {
		action.Noticef("Development: https://%s", image)
	}

	action.SetOutput("release_images", strings.Join(result.releaseImages, ","))
	action.SetOutput("development_images", strings.Join(result.developmentImages, ","))

	images := make([]string, 0, len(result.releaseImages)+len(result.developmentImages))
	images = append(images, result.releaseImages...)
	images = append(images, result.developmentImages...)
	action.SetOutput("images", strings.Join(images, ","))
}

type result struct {
	releaseImages     []string
	developmentImages []string
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func extract(action *githubactions.Action) (*result, error) {
	// extract owner and name to support GitHub forks
	parts := strings.Split(strings.ToLower(action.Getenv("GITHUB_REPOSITORY")), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to extract owner or name")
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

		res := &result{
			developmentImages: []string{
				fmt.Sprintf("ghcr.io/%s/%s-dev:pr-%s", owner, name, branch),
			},
		}

		// no forks, no other repos for Docker Hub
		if owner == "ferretdb" && name == "ferretdb" {
			res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:pr-%s", branch))
		}

		return res, nil

	case "push", "schedule", "workflow_run":
		refType := strings.ToLower(action.Getenv("GITHUB_REF_TYPE"))
		refName := strings.ToLower(action.Getenv("GITHUB_REF_NAME"))

		switch refType {
		case "branch":
			// build on pull_request/pull_request_target for other branches
			if refName != "main" {
				return nil, fmt.Errorf("unhandled branch %q", refName)
			}

			res := &result{
				developmentImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, refName),
				},
			}

			// no forks, no other repos for Docker Hub
			if owner == "ferretdb" && name == "ferretdb" {
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", refName))
			}

			return res, nil

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

			res := &result{
				releaseImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s:%s", owner, name, version),
				},
				developmentImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, version),
				},
			}

			if prerelease == "" {
				res.releaseImages = append(res.releaseImages, fmt.Sprintf("ghcr.io/%s/%s:latest", owner, name))
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/%s-dev:latest", owner, name))
			}

			// no forks, no other repos for Docker Hub
			if owner == "ferretdb" && name == "ferretdb" {
				res.releaseImages = append(res.releaseImages, fmt.Sprintf("ferretdb/ferretdb:%s", version))
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", version))

				if prerelease == "" {
					res.releaseImages = append(res.releaseImages, "ferretdb/ferretdb:latest")
					res.developmentImages = append(res.developmentImages, "ferretdb/ferretdb-dev:latest")
				}
			}

			return res, nil

		default:
			return nil, fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		return nil, fmt.Errorf("unhandled event type %q", event)
	}
}
