package info

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestNewClientWithDiscovery_Success(t *testing.T) {
	var bootstrapClientID, seenAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/get_info.json" {
			t.Errorf("path = %q, want /api/get_info.json", r.URL.Path)
		}
		bootstrapClientID = r.URL.Query().Get("client_id")
		seenAPIKey = r.URL.Query().Get("apikey")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful.",
			"return": {
				"client_id": "979387",
				"contact_id": null,
				"modules": ["Server", "DNS", "Mail"],
				"roles": ["sitehost"]
			}
		}`)
	}))
	defer server.Close()

	c, err := NewClientWithDiscovery(context.Background(), "the-key", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClientWithDiscovery: %v", err)
	}
	if c.ClientID != "979387" {
		t.Errorf("returned ClientID = %q, want 979387", c.ClientID)
	}
	if c.APIKey != "the-key" {
		t.Errorf("returned APIKey = %q, want the-key", c.APIKey)
	}
	if bootstrapClientID != "0" {
		t.Errorf("bootstrap client_id query = %q, want 0 (placeholder)", bootstrapClientID)
	}
	if seenAPIKey != "the-key" {
		t.Errorf("apikey query = %q", seenAPIKey)
	}
}

func TestNewClientWithDiscovery_EmptyKey(t *testing.T) {
	_, err := NewClientWithDiscovery(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "apiKey must not be empty") {
		t.Errorf("err = %v, want apiKey must not be empty", err)
	}
}

func TestNewClientWithDiscovery_StatusFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": false, "msg": "Invalid API key"}`)
	}))
	defer server.Close()

	_, err := NewClientWithDiscovery(context.Background(), "bad-key", api.SetBaseURL(server.URL))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "api/get_info.json") {
		t.Errorf("err = %v, want wrapped api/get_info.json error", err)
	}
}

func TestNewClientWithDiscovery_EmptyClientID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful.",
			"return": {"client_id": "", "modules": [], "roles": []}
		}`)
	}))
	defer server.Close()

	_, err := NewClientWithDiscovery(context.Background(), "the-key", api.SetBaseURL(server.URL))
	if err == nil || !strings.Contains(err.Error(), "empty client_id") {
		t.Errorf("err = %v, want empty client_id", err)
	}
}
