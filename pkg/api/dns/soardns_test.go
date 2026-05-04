package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func formBody(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	v, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return v
}

func TestUpdateSOA_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/update_soa.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("domain") != "example.nz" || v.Get("ns") != "ns1.sitehost.co.nz" {
			t.Errorf("body = %v", v)
		}
		if v.Get("refresh") != "3600" || v.Get("minimum") != "300" {
			t.Errorf("ttls = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpdateSOA(context.Background(), UpdateSOARequest{
		Domain: "example.nz", NS: "ns1.sitehost.co.nz",
		Email:   "support@sitehost.co.nz",
		Refresh: 3600, Retry: 600, Expire: 86400, Minimum: 300,
	})
	if err != nil {
		t.Fatalf("UpdateSOA: %v", err)
	}
	if !got.Status {
		t.Errorf("Status = %v, msg=%q", got.Status, got.Msg)
	}
}

func TestUpdateReverseDNS_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/update_reverse_dns.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("ip_addr") != "192.168.1.105" || v.Get("rdns") != "host.example.com" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateReverseDNS(context.Background(), UpdateReverseDNSRequest{
		IPAddr: "192.168.1.105", RDNS: "host.example.com",
	}); err != nil {
		t.Fatalf("UpdateReverseDNS: %v", err)
	}
}

func TestResetReverseDNS_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/reset_reverse_dns.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("ip_addr") != "192.168.1.105" {
			t.Errorf("ip_addr = %q", v.Get("ip_addr"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).ResetReverseDNS(context.Background(), ResetReverseDNSRequest{
		IPAddr: "192.168.1.105",
	}); err != nil {
		t.Fatalf("ResetReverseDNS: %v", err)
	}
}
