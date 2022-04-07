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
}

type result struct {
	owner string // ferretdb
	name  string // github-actions-dev
	tag   string // pr-add-features or 0.0.1
	ghcr  string // ghcr.io/ferretdb/github-actions-dev:pr-add-features or ghcr.io/ferretdb/github-actions-dev:0.0.1
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with leading `v` and without metadata we don't currently use
var semVerTag = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)

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

	// change name for dance repo
	switch result.name {
	case "dance":
		result.name = "ferretdb"
	case "ferretdb":
		// nothing
	default:
		return nil, fmt.Errorf("unhandled repo %q", repo)
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
			if match == nil || len(match) != 5 {
				return nil, fmt.Errorf("unexpected git tag %q", refName)
			}
			result.name += "-dev" // TODO remove for https://github.com/FerretDB/FerretDB/issues/70
			result.tag = fmt.Sprintf("%s.%s.%s-%s", match[1], match[2], match[3], match[4])

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
