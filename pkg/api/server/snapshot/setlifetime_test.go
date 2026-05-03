package snapshot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestSetLifetime_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("snapshot"); got != "91695" {
			t.Errorf("snapshot = %q", got)
		}
		if got := r.Form.Get("lifetime"); got != "48" {
			t.Errorf("lifetime = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).SetLifetime(context.Background(), SetLifetimeOptions{
		Name: "ch-foo", Snapshot: "91695", Lifetime: 48,
	}); err != nil {
		t.Fatalf("SetLifetime: %v", err)
	}
}

func TestSetLifetime_RequiredFields(t *testing.T) {
	c, _ := api.New("k", "1")
	if _, err := New(c).SetLifetime(context.Background(), SetLifetimeOptions{Snapshot: "1", Lifetime: 1}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("missing Name: err = %v", err)
	}
	if _, err := New(c).SetLifetime(context.Background(), SetLifetimeOptions{Name: "n", Lifetime: 1}); err == nil ||
		!strings.Contains(err.Error(), "Snapshot is required") {
		t.Errorf("missing Snapshot: err = %v", err)
	}
	if _, err := New(c).SetLifetime(context.Background(), SetLifetimeOptions{Name: "n", Snapshot: "1"}); err == nil ||
		!strings.Contains(err.Error(), "Lifetime must be > 0") {
		t.Errorf("missing Lifetime: err = %v", err)
	}
}
