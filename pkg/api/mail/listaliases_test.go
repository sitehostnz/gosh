package mail

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListAliases_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_aliases.json" {
			t.Errorf("path = %q, want /mail/list_aliases.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("domain"); got != "example.co.nz" {
			t.Errorf("domain = %q, want example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{"source": "info@example.co.nz",  "destination": "alice@example.co.nz"},
				{"source": "sales@example.co.nz", "destination": "bob@example.co.nz"}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).ListAliases(context.Background(), ListAliasesOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
	})
	if err != nil {
		t.Fatalf("ListAliases: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	if got.Return[0].Source != "info@example.co.nz" {
		t.Errorf("Return[0].Source = %q", got.Return[0].Source)
	}
	if got.Return[1].Destination != "bob@example.co.nz" {
		t.Errorf("Return[1].Destination = %q", got.Return[1].Destination)
	}
}

func TestListAliases_DomainRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	_, err := New(c).ListAliases(context.Background(), ListAliasesOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error for empty Domain, got nil")
	}
	if !strings.Contains(err.Error(), "Domain is required") {
		t.Errorf("error = %q", err.Error())
	}
}
