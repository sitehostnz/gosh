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

func TestSearchForwards_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/search_forwards.json" {
			t.Errorf("path = %q, want /mail/search_forwards.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("query[destination]"); got != "remote@elsewhere.example" {
			t.Errorf("query[destination] = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{"client_id": "1234", "source": "external@example.co.nz", "destination": "remote@elsewhere.example"}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).SearchForwards(context.Background(), SearchForwardsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Destination:   "remote@elsewhere.example",
	})
	if err != nil {
		t.Fatalf("SearchForwards: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].Source != "external@example.co.nz" {
		t.Errorf("Return[0].Source = %q", got.Return[0].Source)
	}
}

func TestSearchForwards_FilterRequired(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).SearchForwards(context.Background(), SearchForwardsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error for no filters, got nil")
	}
	if !strings.Contains(err.Error(), "at least one filter") {
		t.Errorf("error = %q", err.Error())
	}
}
