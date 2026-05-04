package template

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// TestList_Success exercises the happy path: a well-formed GET request
// with auth params attached, a canned API response, and the parsed
// response surfaced as DomainTemplate values.
//
// Doubles as the test pattern reference for pkg/api/* contributions —
// httptest.Server stands in for api.sitehost.nz, api.SetBaseURL points
// the client at it, request shape and response parsing are both asserted.
func TestList_Success(t *testing.T) {
	t.Parallel()
	const (
		apiKey   = "test-key"
		clientID = "1234"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/dns/domain_templates/list_templates.json" {
			t.Errorf("path = %q, want /dns/domain_templates/list_templates.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("apikey"); got != apiKey {
			t.Errorf("apikey = %q, want %q", got, apiKey)
		}
		if got := r.URL.Query().Get("client_id"); got != clientID {
			t.Errorf("client_id = %q, want %q", got, clientID)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "OK",
			"return": [
				{"client_id": "1234", "template_id": "10", "template_name": "default-nz", "domain_count": "42"},
				{"client_id": "1234", "template_id": "11", "template_name": "fastmail-mx",  "domain_count": "7"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New(apiKey, clientID, api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if !got.Status {
		t.Errorf("Status = false, want true")
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	want := DomainTemplate{
		ClientID: "1234", TemplateID: "10", TemplateName: "default-nz", DomainCount: "42",
	}
	if got.Return[0] != want {
		t.Errorf("Return[0] = %+v, want %+v", got.Return[0], want)
	}
}

// TestList_APIError verifies that a non-success API response (status:false)
// surfaces as an error, not as a silently-empty success.
func TestList_APIError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": false, "msg": "Permission denied"}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).List(context.Background())
	if err == nil {
		t.Fatal("List: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Permission denied") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "Permission denied")
	}
}
