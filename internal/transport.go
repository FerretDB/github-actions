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

package internal

import (
	"net/http"
	"net/http/httputil"

	"github.com/sethvargo/go-githubactions"
)

// transport is an http.RoundTripper with debug logging.
type transport struct {
	t      http.RoundTripper
	action *githubactions.Action
}

// NewTransport returns a new http.RoundTripper that wraps the source with debug logging.
func NewTransport(source http.RoundTripper, action *githubactions.Action) http.RoundTripper {
	return &transport{
		t:      source,
		action: action,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	b, dumpErr := httputil.DumpRequestOut(req, true)
	if dumpErr != nil {
		t.action.Fatalf("%s", dumpErr)
		return nil, dumpErr
	}

	t.action.Debugf("Request:\n%s", b)

	resp, err := t.t.RoundTrip(req)
	if resp == nil {
		return nil, err
	}

	b, dumpErr = httputil.DumpResponse(resp, true)
	if dumpErr != nil {
		t.action.Fatalf("%s", dumpErr)
		return nil, dumpErr
	}

	t.action.Debugf("Response:\n%s", b)

	return resp, err
}
