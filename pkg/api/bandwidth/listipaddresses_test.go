package bandwidth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListIPAddresses_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bandwidth/get_ip_list.json" {
			t.Errorf("path = %q, want /bandwidth/get_ip_list.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"203.0.113.10/32": {
					"ip_addr": "203.0.113.10/32",
					"netmask": "255.255.255.0",
					"prefix": "32",
					"reserved": "0",
					"rdns": "rdns.203.0.113.10.example",
					"addr_family": "4",
					"date_allocated": "2026-03-01 12:00:00"
				}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListIPAddresses(context.Background())
	if err != nil {
		t.Fatalf("ListIPAddresses: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	ip, ok := got.Return["203.0.113.10/32"]
	if !ok {
		t.Fatal("expected 203.0.113.10/32 key")
	}
	if ip.Family != "4" {
		t.Errorf("Family = %q, want 4", ip.Family)
	}
}

func TestListIPAddresses_EmptyAccount(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":[]}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListIPAddresses(context.Background())
	if err != nil {
		t.Fatalf("ListIPAddresses: %v", err)
	}
	if got.Return == nil {
		t.Fatal("Return is nil, want non-nil empty map")
	}
	if len(got.Return) != 0 {
		t.Errorf("len(Return) = %d, want 0", len(got.Return))
	}
}
