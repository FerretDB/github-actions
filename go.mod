module github.com/FerretDB/github-actions

// Our Go actions are limited to version 1.17 as that's the default version on GitHub runners,
// and we don't want to require a separate step of installing Go.
go 1.17

require (
	github.com/google/go-github/v42 v42.0.0
	github.com/sethvargo/go-githubactions v1.0.0
	github.com/stretchr/testify v1.7.1
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sethvargo/go-envconfig v0.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
