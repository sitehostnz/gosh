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

func TestGetServerInfo_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/get_server_info.json" {
			t.Errorf("path = %q, want /mail/get_server_info.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != "sth-mail-air" {
			t.Errorf("server_name = %q, want sth-mail-air", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": {
				"hostname": "mx1.sitehost.co.nz",
				"webmail_url": "https://webmail.sitehost.co.nz",
				"date_added": "0000-00-00 00:00:00",
				"date_updated": "2025-09-14 23:20:51"
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetServerInfo(context.Background(),
		GetServerInfoOptions{ServerOptions: ServerOptions{ServerName: "sth-mail-air"}})
	if err != nil {
		t.Fatalf("GetServerInfo: %v", err)
	}

	if got.Return.Hostname != "mx1.sitehost.co.nz" {
		t.Errorf("Hostname = %q, want mx1.sitehost.co.nz", got.Return.Hostname)
	}
	if got.Return.WebmailURL != "https://webmail.sitehost.co.nz" {
		t.Errorf("WebmailURL = %q", got.Return.WebmailURL)
	}
}

func TestGetServerInfo_ServerNameRequired(t *testing.T) {
	t.Parallel()
	c, err := api.New("k", "1")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).GetServerInfo(context.Background(), GetServerInfoOptions{})
	if err == nil {
		t.Fatal("expected error for empty ServerName, got nil")
	}
	if !strings.Contains(err.Error(), "ServerName is required") {
		t.Errorf("error = %q, want it to contain 'ServerName is required'", err.Error())
	}
}

// TestGetServerInfo_DefaultServerNameInherited locks in the
// NewForServer behaviour: a captured default is used when the
// per-call options omit ServerName.
func TestGetServerInfo_DefaultServerNameInherited(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("server_name"); got != "captured-default" {
			t.Errorf("server_name = %q, want captured-default", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"hostname":"x","webmail_url":"y","date_added":"","date_updated":""}}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	// No ServerName in opt → falls back to captured default.
	if _, err := NewForServer(c, "captured-default").GetServerInfo(
		context.Background(), GetServerInfoOptions{}); err != nil {
		t.Fatalf("GetServerInfo (default): %v", err)
	}
}

// TestGetServerInfo_PerCallOverridesDefault locks in that an
// explicit ServerName in the per-call options beats a captured
// default.
func TestGetServerInfo_PerCallOverridesDefault(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("server_name"); got != "per-call" {
			t.Errorf("server_name = %q, want per-call (per-call must override default)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"hostname":"x","webmail_url":"y","date_added":"","date_updated":""}}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	if _, err := NewForServer(c, "captured-default").GetServerInfo(
		context.Background(), GetServerInfoOptions{ServerOptions: ServerOptions{ServerName: "per-call"}}); err != nil {
		t.Fatalf("GetServerInfo (override): %v", err)
	}
}
