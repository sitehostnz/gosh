package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListLocations_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/list_locations.json" {
			t.Errorf("path = %q, want /server/list_locations.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{
					"public": "1",
					"os": ["Windows", "Linux"],
					"label": "NZ - AKL01",
					"code": "WINCITY",
					"datacenter": "NZ-AKL1",
					"available_ips": 183,
					"available_ipv4": 39,
					"available_ipv6": 144,
					"ipv6": true,
					"public_private_cloud": false,
					"product_types": ["DISK", "UPGRAD", "VDSERV"]
				}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListLocations(context.Background())
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	loc := got.Return[0]
	if loc.Code != "WINCITY" {
		t.Errorf("Code = %q, want WINCITY", loc.Code)
	}
	if loc.AvailableIPs != 183 {
		t.Errorf("AvailableIPs = %d, want 183", loc.AvailableIPs)
	}
	if !loc.IPv6 {
		t.Errorf("IPv6 = false, want true")
	}
	if len(loc.ProductTypes) != 3 {
		t.Errorf("len(ProductTypes) = %d, want 3", len(loc.ProductTypes))
	}
}
