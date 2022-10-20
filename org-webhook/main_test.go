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

			getEnv := testutil.GetEnvFunc(t, map[string]string{
				"GITHUB_SECRET_KEY": tc.secret,
			})
			action := githubactions.New(githubactions.WithGetenv(getEnv))

			h := newWebhookHandler(action)
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
