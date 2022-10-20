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
	"github.com/google/go-github/v45/github"
	"github.com/sethvargo/go-githubactions"
	"log"
	"net/http"
)

// main starts http server. It requires env vars.
// Example:
// WEBHOOK_ADDR="localhost:8088" GITHUB_SECRET_KEY={{WebhookSecret}} go run org-webhook/main.go
func main() {
	action := githubactions.New()

	h := newWebhookHandler(action)
	addr := action.Getenv("WEBHOOK_ADDR")
	http.HandleFunc("/webhook", h.handleWebhook)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// webhookHandler contains secret key.
type webhookHandler struct {
	secretKey []byte
}

// newWebhookHandler creates a handler with secret from env var.
func newWebhookHandler(action *githubactions.Action) *webhookHandler {
	secretKey := action.Getenv("GITHUB_SECRET_KEY")
	return &webhookHandler{
		secretKey: []byte(secretKey),
	}
}

// handleWebhook checks secret and signature, then logs projects_v2_item event.
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#webhook-payload-object-35
func (h *webhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		log.Printf("not valid content type: %s", contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// checks signature then checks secret
	payload, err := github.ValidatePayload(r, h.secretKey)
	if err != nil {
		log.Printf("cannot validate payload: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// log project v2 item event
	if github.WebHookType(r) == "projects_v2_item" {
		// dump the payload
		log.Printf("projects_v2_item: %v", string(payload))
	} else {
		log.Printf("%s event received", github.WebHookType(r))
	}
}
