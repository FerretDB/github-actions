package internal

import (
	"os"
	"sort"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// DebugEnv logs all environment variables that start with `GITHUB_` or `INPUT_`
// in debug level.
func DebugEnv(action *githubactions.Action) {
	res := make([]string, 0, 30)
	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") || strings.HasPrefix(l, "INPUT_") {
			res = append(res, l)
		}
	}

	sort.Strings(res)

	action.Debugf("Dumping environment variables:")
	for _, l := range res {
		action.Debugf("\t%s", l)
	}
}
