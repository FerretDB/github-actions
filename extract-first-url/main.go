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
	"bufio"
	"flag"
	"os"
	"regexp"

	"github.com/FerretDB/github-actions/internal"
	"github.com/sethvargo/go-githubactions"
)

const (
	regexMatchURL  = `(https?://[^\s]+)`
	emptyString    = ""
	outputURLParam = "deployment_url"
	tempFilepath   = "./deploy.txt"
)

func main() {

	flag.Parse()
	action := githubactions.New()
	internal.DebugEnv(action)

	//open a file
	file, err := os.Open(tempFilepath)
	if err != nil {
		action.Fatalf("%s", err)
	}
	defer file.Close()

	//get first url form file
	url := getFirstURLFromFile(file)

	//set url as output parameter
	if url != emptyString {
		action.SetOutput(outputURLParam, url)
	}
}

func getFirstURLFromFile(inputFile *os.File) string {
	urlPattern := regexp.MustCompile(regexMatchURL)
	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		url := urlPattern.FindString(scanner.Text())
		if url != emptyString {
			return url
		}
	}
	return emptyString
}
