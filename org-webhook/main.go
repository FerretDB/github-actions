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
	"github.com/gorilla/mux"
	"github.com/sethvargo/go-githubactions"
	"log"
	"net/http"
	"time"

	_ "github.com/FerretDB/github-actions/internal"
)

func main() {
	action := githubactions.New()

	h := newWebhookHandler(action)

	r := mux.NewRouter()
	r.HandleFunc("/webhook", h.handleWebhook).
		Methods("POST").
		Headers("Content-Type", "application/json")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8088",
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// webhookHandler contains secret
type webhookHandler struct {
	secretKey []byte
}

// newWebhookHandler creates a webhookhandler
func newWebhookHandler(action *githubactions.Action) *webhookHandler {
	secretKey := action.Getenv("GITHUB_SECRET_KEY")
	return &webhookHandler{
		secretKey: []byte(secretKey),
	}
}

// handleWebhook checks secret and signature and logs projects_v2_item event
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#webhook-payload-object-35
func (h *webhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// check secret
	payload, err := github.ValidatePayload(r, h.secretKey)
	if err != nil {
		log.Printf("cannot validate payload: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// check signature
	signature := "" // TODO: populate this
	err = github.ValidateSignature(signature, payload, h.secretKey)
	if err != nil {
		log.Printf("cannot validate signature: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("cannot parse payload: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case *github.ProjectEvent:
		log.Printf("Project Event: %#v", event)
	case *github.PingEvent:
		log.Print("Ping Event: %#v", event)
	default:
		log.Print("Unexpected event type %T", event)
	}
}
