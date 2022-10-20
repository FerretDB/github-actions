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
	"bytes"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/projectv2-item-edited.json
var edited []byte

func TestRun(t *testing.T) {
	// env vars are set and cleaned up by tests, so they do not run in parallel.
	t.Run("missing WEBHOOK_ADDR", func(t *testing.T) {
		t.Setenv("GITHUB_SECRET_KEY", "superSecret")
		err := run()
		assert.EqualError(t, err, "missing WEBHOOK_ADDR env var")
	})

	t.Run("missing GITHUB_SECRET_KEY", func(t *testing.T) {
		t.Setenv("WEBHOOK_ADDR", "localhost:8088")
		err := run()
		assert.EqualError(t, err, "missing GITHUB_SECRET_KEY env var")
	})
}

func TestWebhook(t *testing.T) {
	t.Parallel()

	secret := "verySecret"
	signature := "sha256=95f148f147729be0803d90ada4478ccc61342a8d788a2081b16bacd02cad2d59"

	testcases := []struct {
		name        string
		contentType string
		method      string
		secret      string
		signature   string
		statusCode  int
		r           []byte
	}{{
		name:        "edited",
		contentType: "application/json",
		method:      http.MethodPost,
		secret:      secret,
		signature:   signature,
		statusCode:  http.StatusOK,
	}, {
		name:        "wrong secret",
		contentType: "application/json",
		method:      http.MethodPost,
		secret:      "not valid secret",
		signature:   signature,
		statusCode:  http.StatusUnauthorized,
	}, {
		name:        "wrong signature",
		contentType: "application/json",
		method:      http.MethodPost,
		secret:      secret,
		signature:   "sha256=fb2fa31da7ad769064cfb975384a032a1750ebaa4f8f3cacf564932beb707a70",
		statusCode:  http.StatusUnauthorized,
	}, {
		name:        "empty content type",
		contentType: "",
		method:      http.MethodPost,
		secret:      secret,
		signature:   signature,
		statusCode:  http.StatusBadRequest,
	}, {
		name:        "http get method",
		contentType: "application/json",
		method:      http.MethodGet,
		secret:      secret,
		signature:   signature,
		statusCode:  http.StatusNotFound,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newWebhookHandler(tc.secret)

			reader := bytes.NewReader(edited)
			req := httptest.NewRequest(tc.method, "/webhook", reader)
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("x-hub-signature-256", tc.signature)

			w := httptest.NewRecorder()
			h.handleWebhook(w, req)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, res.StatusCode, tc.statusCode)
		})
	}
}
