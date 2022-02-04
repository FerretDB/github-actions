package internal

import (
	"os"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// DumpEnv logs all environment variables that start with `GITHUB_`.
func DumpEnv(action *githubactions.Action) {
	action.Infof("Dumping environment variables:")

	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") {
			action.Infof("\t%s", l)
		}
	}
}
