package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sign(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func pushBody(repo, ref string) string {
	return fmt.Sprintf(`{"ref":%q,"repository":{"name":%q}}`, ref, repo)
}

func TestVerifySignature(t *testing.T) {
	secret = "test-secret"
	const body = "hello world"
	valid := sign(secret, body)

	cases := []struct {
		name string
		sig  string
		want bool
	}{
		{"valid", valid, true},
		{"empty", "", false},
		{"malformed", "sha256=00", false},
		{"wrong secret", sign("other-secret", body), false},
		{"missing prefix", strings.TrimPrefix(valid, "sha256="), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := verifySignature([]byte(body), tc.sig); got != tc.want {
				t.Errorf("verifySignature(%q) = %v, want %v", tc.sig, got, tc.want)
			}
		})
	}
}

func TestWebhook(t *testing.T) {
	const testSecret = "test-secret"
	secret = testSecret
	config = Config{Projects: map[string]Project{
		"pinned":    {Path: "/tmp/pinned", Branch: "main"},
		"anybranch": {Path: "/tmp/anybranch"}, // no branch pinned -> accept any
	}}

	orig := rebuildFn
	t.Cleanup(func() { rebuildFn = orig })

	cases := []struct {
		name       string
		body       string
		event      string
		sign       bool
		rawSig     string
		wantStatus int
		wantRepo   string
	}{
		{
			name: "valid push triggers rebuild", sign: true,
			body: pushBody("pinned", "refs/heads/main"),
			wantStatus: http.StatusAccepted, wantRepo: "pinned",
		},
		{
			name: "unpinned project accepts any branch", sign: true,
			body: pushBody("anybranch", "refs/heads/feature"),
			wantStatus: http.StatusAccepted, wantRepo: "anybranch",
		},
		{
			name: "non-matching branch ignored", sign: true,
			body: pushBody("pinned", "refs/heads/dev"),
			wantStatus: http.StatusOK, wantRepo: "",
		},
		{
			name: "unknown repo ignored", sign: true,
			body: pushBody("ghost", "refs/heads/main"),
			wantStatus: http.StatusOK, wantRepo: "",
		},
		{
			name: "missing signature rejected", sign: false,
			body: pushBody("pinned", "refs/heads/main"),
			wantStatus: http.StatusUnauthorized, wantRepo: "",
		},
		{
			name: "invalid signature rejected", rawSig: "sha256=deadbeef",
			body: pushBody("pinned", "refs/heads/main"),
			wantStatus: http.StatusUnauthorized, wantRepo: "",
		},
		{
			name: "ping event acknowledged", sign: true, event: "ping",
			body: "{}", wantStatus: http.StatusOK, wantRepo: "",
		},
		{
			name: "invalid json rejected", sign: true,
			body: "not json", wantStatus: http.StatusBadRequest, wantRepo: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotRepo string
			rebuildFn = func(repo string, _ Project) { gotRepo = repo }

			req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(tc.body))
			if tc.sign {
				req.Header.Set("X-Hub-Signature-256", sign(testSecret, tc.body))
			}
			if tc.rawSig != "" {
				req.Header.Set("X-Hub-Signature-256", tc.rawSig)
			}
			if tc.event != "" {
				req.Header.Set("X-GitHub-Event", tc.event)
			}

			rec := httptest.NewRecorder()
			webhook(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if gotRepo != tc.wantRepo {
				t.Errorf("rebuild triggered for %q, want %q", gotRepo, tc.wantRepo)
			}
		})
	}
}
