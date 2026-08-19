package api

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// TestTransportErrorRedacted — timeouts, DNS failures, TLS errors and
// connection resets all return *url.Error, whose Error() embeds the full
// request URL. net/http strips userinfo passwords, never query
// parameters, so without redaction every transport failure logs the
// caller's credential. Uses an unroutable address so the dial fails
// without any network dependency.
func TestTransportErrorRedacted(t *testing.T) {
	t.Parallel()
	// 127.0.0.1:1 refuses the connection immediately — a fast, offline,
	// deterministic transport failure with the default http.Client.
	c, err := New("SUPERSECRETKEY", "1", SetBaseURL("https://127.0.0.1:1"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req, err := c.NewRequest(http.MethodGet, "vnc/valid_token.json?token=abc", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	callErr := c.Do(context.Background(), req, nil)
	if callErr == nil {
		t.Fatal("expected a transport error")
	}
	msg := callErr.Error()
	if strings.Contains(msg, "SUPERSECRETKEY") {
		t.Errorf("transport error leaks the api key: %s", msg)
	}
	if !strings.Contains(msg, "apikey=REDACTED") {
		t.Errorf("transport error should carry the redacted marker: %s", msg)
	}
	// The request itself must still hold the real key — redaction is for
	// display, not for the request.
	if got := req.URL.Query().Get("apikey"); got != "SUPERSECRETKEY" {
		t.Errorf("request URL was mutated: apikey=%q", got)
	}
}
