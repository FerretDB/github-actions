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
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/sethvargo/go-githubactions"

	"github.com/FerretDB/github-actions/internal"
)

func main() {
	flag.Parse()

	action := githubactions.New()

	internal.DebugEnv(action)

	result, err := extract(action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	setResults(action, result)
}

type result struct {
	allInOneImages    []string
	developmentImages []string
	productionImages  []string
}

// Sort sorts all images in-place.
func (r *result) Sort() {
	sort.Strings(r.allInOneImages)
	sort.Strings(r.developmentImages)
	sort.Strings(r.productionImages)
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

//nolint:goconst // "ferretdb" means different things
func extract(getenv githubactions.GetenvFunc) (*result, error) {
	// extract owner and name to support GitHub forks
	parts := strings.Split(strings.ToLower(getenv("GITHUB_REPOSITORY")), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to extract owner or name")
	}
	owner := parts[0]
	name := parts[1]

	// extract tags for various events
	event := getenv("GITHUB_EVENT_NAME")
	switch event {
	case "pull_request", "pull_request_target":
		// for branches like "dependabot/submodules/XXX"
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		parts = strings.Split(branch, "/")
		branch = parts[len(parts)-1]

		res := &result{
			developmentImages: []string{
				fmt.Sprintf("ghcr.io/%s/%s-dev:pr-%s", owner, name, branch),
			},
		}

		// all-in-one only for FerretDB
		if name == "ferretdb" {
			res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ghcr.io/%s/all-in-one:pr-%s", owner, branch))

			// no forks, no other repos for Quay.io and Docker Hub
			if owner == "ferretdb" {
				res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("quay.io/ferretdb/all-in-one:pr-%s", branch))
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:pr-%s", branch))

				res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ferretdb/all-in-one:pr-%s", branch))
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:pr-%s", branch))
			}
		}

		res.Sort()
		return res, nil

	case "push", "schedule", "workflow_run":
		refType := strings.ToLower(getenv("GITHUB_REF_TYPE"))
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType {
		case "branch":
			// build on pull_request/pull_request_target for other branches
			switch {
			case refName == "main":
				// nothing
			case strings.HasPrefix(refName, "releases/"):
				refName = strings.ReplaceAll(refName, "/", "-")
			default:
				return nil, fmt.Errorf("unhandled branch %q", refName)
			}

			res := &result{
				developmentImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, refName),
				},
			}

			// all-in-one only for FerretDB
			if name == "ferretdb" {
				res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ghcr.io/%s/all-in-one:%s", owner, refName))

				// no forks, no other repos for Quay.io and Docker Hub
				if owner == "ferretdb" {
					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("quay.io/ferretdb/all-in-one:%s", refName))
					res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:%s", refName))

					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ferretdb/all-in-one:%s", refName))
					res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", refName))
				}
			}

			res.Sort()
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
				developmentImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, version),
				},
				productionImages: []string{
					fmt.Sprintf("ghcr.io/%s/%s:%s", owner, name, version),
				},
			}

			// all-in-one only for FerretDB
			if name == "ferretdb" {
				res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ghcr.io/%s/all-in-one:%s", owner, version))

				// no forks, no other repos for Quay.io and Docker Hub
				if owner == "ferretdb" {
					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("quay.io/ferretdb/all-in-one:%s", version))
					res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:%s", version))
					res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/ferretdb:%s", version))

					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ferretdb/all-in-one:%s", version))
					res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", version))
					res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/ferretdb:%s", version))
				}
			}

			if prerelease == "" {
				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, name, major))
				res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/%s:%s", owner, name, major))

				res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/%s-dev:latest", owner, name))
				res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/%s:latest", owner, name))

				// all-in-one only for FerretDB
				if name == "ferretdb" {
					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ghcr.io/%s/all-in-one:%s", owner, major))

					res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ghcr.io/%s/all-in-one:latest", owner))

					// no forks, no other repos for Quay.io and Docker Hub
					//nolint:lll // for readibility
					if owner == "ferretdb" {
						res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("quay.io/ferretdb/all-in-one:%s", major))
						res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:%s", major))
						res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/ferretdb:%s", major))

						res.allInOneImages = append(res.allInOneImages, "quay.io/ferretdb/all-in-one:latest")
						res.developmentImages = append(res.developmentImages, "quay.io/ferretdb/ferretdb-dev:latest")
						res.productionImages = append(res.productionImages, "quay.io/ferretdb/ferretdb:latest")

						res.allInOneImages = append(res.allInOneImages, fmt.Sprintf("ferretdb/all-in-one:%s", major))
						res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", major))
						res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/ferretdb:%s", major))

						res.allInOneImages = append(res.allInOneImages, "ferretdb/all-in-one:latest")
						res.developmentImages = append(res.developmentImages, "ferretdb/ferretdb-dev:latest")
						res.productionImages = append(res.productionImages, "ferretdb/ferretdb:latest")
					}
				}
			}

			res.Sort()
			return res, nil

		default:
			return nil, fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		return nil, fmt.Errorf("unhandled event type %q", event)
	}
}

// setResults sets action output parameters, summary, etc.
func setResults(action *githubactions.Action, result *result) {
	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tImage\t\n")
	fmt.Fprintf(w, "\t----\t-----\t\n")

	for _, image := range result.allInOneImages {
		u := imageURL(image)
		action.Noticef("All-in-one: %s (see %s)", image, u)
		fmt.Fprintf(w, "\tAll-in-one\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range result.developmentImages {
		u := imageURL(image)
		action.Noticef("Development: %s (see %s)", image, u)
		fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range result.productionImages {
		u := imageURL(image)
		action.Noticef("Production: %s (see %s)", image, u)
		fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	action.SetOutput("all_in_one_images", strings.Join(result.allInOneImages, ","))
	action.SetOutput("development_images", strings.Join(result.developmentImages, ","))
	action.SetOutput("production_images", strings.Join(result.productionImages, ","))
}

// imageURL returns URL for the given image name.
func imageURL(name string) string {
	switch {
	case strings.HasPrefix(name, "ghcr.io/"):
		return fmt.Sprintf("https://%s", name)
	case strings.HasPrefix(name, "quay.io/"):
		return fmt.Sprintf("https://%s", name)
	}

	name, _, _ = strings.Cut(name, ":")

	// there is not easy way to get Docker Hub URL for the given tag
	return fmt.Sprintf("https://hub.docker.com/r/%s/tags", name)
}
