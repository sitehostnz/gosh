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

func TestListDomains_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_domains.json" {
			t.Errorf("path = %q, want /mail/list_domains.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != "sth-mail-air" {
			t.Errorf("server_name = %q, want sth-mail-air", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{
					"client_id": "1234",
					"domain": "example.co.nz",
					"parent_domain": null,
					"catch_all": "",
					"state": "1",
					"accounts": 5,
					"nicknames": 2,
					"forwarders": 1,
					"total_used": 8
				}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListDomains(context.Background(),
		ListDomainsOptions{ServerOptions: ServerOptions{ServerName: "sth-mail-air"}})
	if err != nil {
		t.Fatalf("ListDomains: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	d := got.Return[0]
	if d.Domain != "example.co.nz" {
		t.Errorf("Domain = %q, want example.co.nz", d.Domain)
	}
	if d.State != "1" {
		t.Errorf("State = %q, want 1", d.State)
	}
	if d.Accounts != 5 {
		t.Errorf("Accounts = %d, want 5", d.Accounts)
	}
	if d.TotalUsed != 8 {
		t.Errorf("TotalUsed = %d, want 8", d.TotalUsed)
	}
}

func TestListDomains_ServerNameRequired(t *testing.T) {
	t.Parallel()
	c, err := api.New("k", "1")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).ListDomains(context.Background(), ListDomainsOptions{})
	if err == nil {
		t.Fatal("expected error for empty ServerName, got nil")
	}
	if !strings.Contains(err.Error(), "ServerName is required") {
		t.Errorf("error = %q, want it to contain 'ServerName is required'", err.Error())
	}
}
