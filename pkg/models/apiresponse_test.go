package models

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func errFor(t *testing.T, rawURL string) *ErrorResponse {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return &ErrorResponse{
		Response: &http.Response{
			StatusCode: 400,
			Request:    &http.Request{Method: http.MethodGet, URL: u},
		},
		Message: "The token is invalid.",
	}
}

// TestErrorRedactsCredentials is the whole point: an unredacted URL puts
// the caller's API key into every error string, and error strings get
// wrapped, logged and forwarded.
func TestErrorRedactsCredentials(t *testing.T) {
	t.Parallel()
	const key = "197dd5a9711737395988c8f00c8343a9"
	e := errFor(t, "https://api.sitehost.nz/1.5/vnc/valid_token.json?apikey="+key+"&client_id=1&token=abc123")

	got := e.Error()
	if strings.Contains(got, key) {
		t.Errorf("error leaks the api key: %s", got)
	}
	// The vnc token is deliberately NOT redacted: it is the subject of
	// the request, short-lived and console-scoped, and it is the field
	// you need when the call fails. Hiding it from the caller who
	// supplied it makes a worse error, not a safer one.
	if !strings.Contains(got, "abc123") {
		t.Errorf("vnc token should survive — it is what failed: %s", got)
	}
	// Still useful for debugging: method, path, status and message survive.
	for _, want := range []string{"GET", "/vnc/valid_token.json", "400", "The token is invalid.", "token=abc123"} {
		if !strings.Contains(got, want) {
			t.Errorf("error dropped %q, leaving little to debug with: %s", want, got)
		}
	}
	if !strings.Contains(got, "client_id=1") {
		t.Errorf("non-credential parameters should survive: %s", got)
	}
}

// TestRedactURLDoesNotMutate — redaction is for display. A caller
// inspecting the request afterwards must still see what was sent.
func TestRedactURLDoesNotMutate(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("https://api.sitehost.nz/1.5/x.json?apikey=secret123&a=1")
	before := u.String()
	_ = RedactURL(u)
	if u.String() != before {
		t.Errorf("RedactURL mutated its argument:\n before %s\n after  %s", before, u.String())
	}
}

// TestErrorWithoutResponse — previously a nil Response panicked. An
// error type that panics while being printed is worse than an
// unhelpful message.
func TestErrorWithoutResponse(t *testing.T) {
	t.Parallel()
	e := &ErrorResponse{Message: "boom"}
	if got := e.Error(); got != "boom" {
		t.Errorf("Error() = %q, want %q", got, "boom")
	}
}

func TestRedactURLNil(t *testing.T) {
	t.Parallel()
	if got := RedactURL(nil); got != "" {
		t.Errorf("RedactURL(nil) = %q, want empty", got)
	}
}
