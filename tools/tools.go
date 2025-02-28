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

package tools

//go:generate go build -v -o ../bin/ github.com/go-task/task/v3/cmd/task
//go:generate go build -v -o ../bin/ github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go build -v -o ../bin/ github.com/quasilyte/go-consistent
//go:generate go build -v -o ../bin/ golang.org/x/perf/cmd/benchstat
//go:generate go build -v -o ../bin/ golang.org/x/tools/cmd/godoc
//go:generate go build -v -o ../bin/ golang.org/x/tools/cmd/goimports
//go:generate go build -v -o ../bin/ golang.org/x/tools/cmd/stringer
//go:generate go build -v -o ../bin/ mvdan.cc/gofumpt
