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

func TestSearchAccounts_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/search_accounts.json" {
			t.Errorf("path = %q, want /mail/search_accounts.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("query[emailaddr]"); got != "alice@example.co.nz" {
			t.Errorf("query[emailaddr] = %q, want alice@example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{
					"client_id": "1234",
					"emailaddr": "alice@example.co.nz",
					"label": "Alice",
					"username": "alice@example.co.nz",
					"autoresponder": "0",
					"autoresponder_text": "",
					"active": "yes",
					"quota": "0",
					"spam_strategy": "0"
				}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).SearchAccounts(context.Background(), SearchAccountsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		EmailAddr:     "alice@example.co.nz",
	})
	if err != nil {
		t.Fatalf("SearchAccounts: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].EmailAddr != "alice@example.co.nz" {
		t.Errorf("EmailAddr = %q", got.Return[0].EmailAddr)
	}
	// Search returns the narrower 9-field shape; quota_used / message_count
	// / last_updated should be empty (zero value).
	if got.Return[0].QuotaUsed != "" {
		t.Errorf("QuotaUsed = %q, want empty (search shape doesn't include it)", got.Return[0].QuotaUsed)
	}
}

func TestSearchAccounts_FilterRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	_, err := New(c).SearchAccounts(context.Background(), SearchAccountsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error for no filters set, got nil")
	}
	if !strings.Contains(err.Error(), "at least one filter") {
		t.Errorf("error = %q, want it to contain 'at least one filter'", err.Error())
	}
}
