package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGenerateNetworkConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/generate_network_config.json" {
			t.Errorf("path = %q, want /server/generate_network_config.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "ch-foo" {
			t.Errorf("name = %q, want ch-foo", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"/etc/netplan/50-cloud-init.yaml": "network:\n    version: 2\n    ethernets:\n        eth0:\n            addresses:\n                - 198.51.100.10/24\n"
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GenerateNetworkConfig(context.Background(), GenerateNetworkConfigOptions{Name: "ch-foo"})
	if err != nil {
		t.Fatalf("GenerateNetworkConfig: %v", err)
	}

	cfg, ok := got.Return["/etc/netplan/50-cloud-init.yaml"]
	if !ok {
		t.Fatal("expected netplan config in Return map")
	}
	if !strings.Contains(cfg, "network:") {
		t.Errorf("config does not contain 'network:': %q", cfg[:min(60, len(cfg))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
