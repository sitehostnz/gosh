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

func TestListAccounts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_accounts.json" {
			t.Errorf("path = %q, want /mail/list_accounts.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != "sth-mail-air" {
			t.Errorf("server_name = %q, want sth-mail-air", got)
		}
		if got := r.URL.Query().Get("domain"); got != "example.co.nz" {
			t.Errorf("domain = %q, want example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{
					"client_id": "1234",
					"username": "alice@example.co.nz",
					"label": "Alice",
					"emailaddr": "alice@example.co.nz",
					"autoresponder": "0",
					"autoresponder_text": "",
					"spam_strategy": "0",
					"active": "yes",
					"quota": "0",
					"quota_used": "1024",
					"quota_percent": "5",
					"message_count": "12",
					"last_updated": "2026-05-01 09:00:00"
				}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListAccounts(context.Background(), ListAccountsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
	})
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	a := got.Return[0]
	if a.EmailAddr != "alice@example.co.nz" {
		t.Errorf("EmailAddr = %q, want alice@example.co.nz", a.EmailAddr)
	}
	if a.QuotaUsed != "1024" {
		t.Errorf("QuotaUsed = %q, want 1024", a.QuotaUsed)
	}
	if a.MessageCount != "12" {
		t.Errorf("MessageCount = %q, want 12", a.MessageCount)
	}
}

func TestListAccounts_DomainRequired(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).ListAccounts(context.Background(), ListAccountsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
	})
	if err == nil {
		t.Fatal("expected error for empty Domain, got nil")
	}
	if !strings.Contains(err.Error(), "Domain is required") {
		t.Errorf("error = %q, want it to contain 'Domain is required'", err.Error())
	}
}

func TestListAccounts_EmailAddrFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filters[emailaddr]"); got != "alice@example.co.nz" {
			t.Errorf("filters[emailaddr] = %q, want alice@example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[]}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).ListAccounts(context.Background(), ListAccountsOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
		EmailAddr:     "alice@example.co.nz",
	}); err != nil {
		t.Fatalf("ListAccounts (filter): %v", err)
	}
}
