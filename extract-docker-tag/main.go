package main

import (
	"flag"

	_ "github.com/sethvargo/go-githubactions"
)

func main() {
	flag.Parse()
}

func extractDockerTag() (string, error) {
	return "", nil
}
