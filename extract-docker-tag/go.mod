module github.com/FerretDB/github-actions/extract-docker-tag

go 1.17

// Until those PRs are merged:
// * https://github.com/sethvargo/go-githubactions/pull/30
// * https://github.com/sethvargo/go-githubactions/pull/31
replace github.com/sethvargo/go-githubactions => github.com/AlekSi/go-githubactions v0.5.1-0.20211229150045-47dd76fc5d73

require (
	github.com/sethvargo/go-githubactions v0.5.0
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
