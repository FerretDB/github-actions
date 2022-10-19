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

//go:embed testdata/webhook-ping.json
var webhook []byte

func TestWebhook(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		r          []byte
		statusCode int
	}{{
		name:       "ping",
		r:          webhook,
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
			req.Header.Set("x-hub-signature-256", "sha256=b6911ff5baf4cdf9d8d2aa73cd253d6b84ec9834ae517dab10d82b293166e86b")

			w := httptest.NewRecorder()
			h.handleWebhook(w, req)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, res.StatusCode, tc.statusCode)
		})
	}
}
