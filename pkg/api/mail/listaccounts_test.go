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
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_accounts.json" {
			t.Errorf("path = %q, want /mail/list_accounts.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != testServerName {
			t.Errorf("server_name = %q, want %s", got, testServerName)
		}
		if got := r.URL.Query().Get("domain"); got != testDomain {
			t.Errorf("domain = %q, want %s", got, testDomain)
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
		ServerOptions: ServerOptions{ServerName: testServerName},
		Domain:        testDomain,
	})
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	a := got.Return[0]
	if a.EmailAddr != testEmail {
		t.Errorf("EmailAddr = %q, want %s", a.EmailAddr, testEmail)
	}
	if a.QuotaUsed != "1024" {
		t.Errorf("QuotaUsed = %q, want 1024", a.QuotaUsed)
	}
	if a.MessageCount != "12" {
		t.Errorf("MessageCount = %q, want 12", a.MessageCount)
	}
}

func TestListAccounts_DomainRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	_, err := New(c).ListAccounts(context.Background(), ListAccountsOptions{
		ServerOptions: ServerOptions{ServerName: testServerName},
	})
	if err == nil {
		t.Fatal("expected error for empty Domain, got nil")
	}
	if !strings.Contains(err.Error(), "Domain is required") {
		t.Errorf("error = %q, want it to contain 'Domain is required'", err.Error())
	}
}

func TestListAccounts_EmailAddrFilter(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filters[emailaddr]"); got != testEmail {
			t.Errorf("filters[emailaddr] = %q, want %s", got, testEmail)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[]}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).ListAccounts(context.Background(), ListAccountsOptions{
		ServerOptions: ServerOptions{ServerName: testServerName},
		Domain:        testDomain,
		EmailAddr:     testEmail,
	}); err != nil {
		t.Fatalf("ListAccounts (filter): %v", err)
	}
}
