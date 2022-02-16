package internal

import (
	"os"
	"sort"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// DumpEnv logs all environment variables that start with `GITHUB_` or `INPUT_`.
func DumpEnv(action *githubactions.Action) {
	res := make([]string, 0, 30)
	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") || strings.HasPrefix(l, "INPUT_") {
			res = append(res, l)
		}
	}

	sort.Strings(res)

	action.Infof("Dumping environment variables:")
	for _, l := range res {
		action.Infof("\t%s", l)
	}
}
