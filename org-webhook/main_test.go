package main

import (
	"bytes"
	_ "embed"
	"github.com/FerretDB/github-actions/internal/testutil"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/projectv2-item-edited.json
var edited []byte

func TestWebhook(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		r          []byte
		statusCode int
	}{{
		name:       "edited",
		r:          edited,
		statusCode: http.StatusOK,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			getEnv := testutil.GetEnvFunc(t, map[string]string{
				"GITHUB_SECRET_KEY": "verySecret",
			})
			action := githubactions.New(githubactions.WithGetenv(getEnv))

			h := newWebhookHandler(action)
			reader := bytes.NewReader(tc.r)
			req := httptest.NewRequest(http.MethodPost, "/webhook", reader)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-hub-signature-256", "sha256=1c9b1e78959626ace6f90171161181acfb16d27c5f243dcd7a07bfb722872eea")

			w := httptest.NewRecorder()
			h.handleWebhook(w, req)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, res.StatusCode, tc.statusCode)
		})
	}
}
