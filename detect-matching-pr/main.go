package main

import (
	"flag"

	_ "github.com/google/go-github/v42/github"
	_ "github.com/sethvargo/go-githubactions"
)

func main() {
	flag.Parse()
}
