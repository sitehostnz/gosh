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

func TestListForwards_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_forwards.json" {
			t.Errorf("path = %q, want /mail/list_forwards.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{"source": "external@example.co.nz", "destination": "remote@elsewhere.example"}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).ListForwards(context.Background(), ListForwardsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
	})
	if err != nil {
		t.Fatalf("ListForwards: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].Destination != "remote@elsewhere.example" {
		t.Errorf("Return[0].Destination = %q", got.Return[0].Destination)
	}
}

func TestListForwards_DomainRequired(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).ListForwards(context.Background(), ListForwardsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Domain is required") {
		t.Errorf("error = %q", err.Error())
	}
}
