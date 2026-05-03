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

func TestSearchAliases_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/search_aliases.json" {
			t.Errorf("path = %q, want /mail/search_aliases.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("query[source]"); got != "info@example.co.nz" {
			t.Errorf("query[source] = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{"client_id": "1234", "source": "info@example.co.nz", "destination": "alice@example.co.nz"}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).SearchAliases(context.Background(), SearchAliasesOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "info@example.co.nz",
	})
	if err != nil {
		t.Fatalf("SearchAliases: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].ClientID != "1234" {
		t.Errorf("Return[0].ClientID = %q", got.Return[0].ClientID)
	}
}

func TestSearchAliases_FilterRequired(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).SearchAliases(context.Background(), SearchAliasesOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error for no filters, got nil")
	}
	if !strings.Contains(err.Error(), "at least one filter") {
		t.Errorf("error = %q", err.Error())
	}
}
