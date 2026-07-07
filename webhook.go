package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

var secret string

type pushPayload struct {
	Ref        string `json:"ref"` // e.g. "refs/heads/main"
	Repository struct {
		Name string `json:"name"` // e.g. "disping"
	} `json:"repository"`
}

func verifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 25<<20)) // 25 MiB
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	if !verifySignature(body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "Missing or invalid signature", http.StatusUnauthorized)
		return
	}

	if event := r.Header.Get("X-GitHub-Event"); event == "ping" {
		log.Printf("Received ping event.")
		reply(w, http.StatusOK, "Pong")
		return
	}

	var payload pushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	repo := payload.Repository.Name
	project, ok := config.Projects[repo]
	if !ok {
		log.Printf("No project configured for repo %q, ignoring", repo)
		reply(w, http.StatusOK, "no project configured for %q\n", repo)
		return
	}

	// Optional branch filter.
	if branch := strings.TrimPrefix(payload.Ref, "refs/heads/"); branch != "" && branch != project.Branch {
		log.Printf("Push to %s/%s ignored (watching %q)", repo, branch, project.Branch)
		reply(w, http.StatusOK, "branch %q ignored\n", branch)
		return
	}

	log.Printf("Push to %q accepted, triggering rebuild", repo)
	startRebuild(repo, project)

	reply(w, http.StatusAccepted, "rebuilding %q\n", repo)
}
