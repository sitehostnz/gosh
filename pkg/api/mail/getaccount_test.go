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

const (
	testServerName = "sth-mail-air"
	testDomain     = "example.co.nz"
	testEmail      = "alice@example.co.nz"
)

func TestGetAccount_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/get_account.json" {
			t.Errorf("path = %q, want /mail/get_account.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != testServerName {
			t.Errorf("server_name = %q, want %s", got, testServerName)
		}
		if got := r.URL.Query().Get("email"); got != "test@example.co.nz" {
			t.Errorf("email = %q, want test@example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": {
				"client_id": "1234",
				"username": "test@example.co.nz",
				"label": "Test Account",
				"emailaddr": "test@example.co.nz",
				"autoresponder": "0",
				"autoresponder_text": "",
				"spam_strategy": "0",
				"active": "yes",
				"quota": "0",
				"quota_used": "0",
				"quota_percent": "0",
				"message_count": "0",
				"last_updated": "0000-00-00 00:00:00",
				"key": "",
				"date_added": "2026-05-02 13:00:00"
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetAccount(context.Background(), GetAccountOptions{
		ServerOptions: ServerOptions{ServerName: testServerName},
		Email:         "test@example.co.nz",
	})
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}

	if got.Return.EmailAddr != "test@example.co.nz" {
		t.Errorf("EmailAddr = %q, want test@example.co.nz", got.Return.EmailAddr)
	}
	if got.Return.Active != "yes" {
		t.Errorf("Active = %q, want yes", got.Return.Active)
	}
	if got.Return.DateAdded != "2026-05-02 13:00:00" {
		t.Errorf("DateAdded = %q, want 2026-05-02 13:00:00", got.Return.DateAdded)
	}
}

func TestGetAccount_EmailRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	_, err := New(c).GetAccount(context.Background(), GetAccountOptions{
		ServerOptions: ServerOptions{ServerName: testServerName},
	})
	if err == nil {
		t.Fatal("expected error for empty Email, got nil")
	}
	if !strings.Contains(err.Error(), "Email is required") {
		t.Errorf("error = %q, want it to contain 'Email is required'", err.Error())
	}
}
