package srs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestWhois_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/srs/whois.json" {
			t.Errorf("path = %q, want /srs/whois.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("domain"); got != "example.co.nz" {
			t.Errorf("domain = %q, want example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"SourceIP": "203.0.113.10",
			"return": {
				"DomainName": "example.co.nz",
				"Status":     "Active",
				"RegisteredDate": "2020-01-01 00:00:00",
				"ModifiedDate":   "2025-04-15 00:00:00",
				"BilledUntil":    "2026-01-01 00:00:00",
				"NameServers": [
					{"FQDN": "ns1.sitehost.nz", "IP4Addr": "203.0.113.1", "IP6Addr": "2001:db8::1"},
					{"FQDN": "ns2.sitehost.nz", "IP4Addr": "203.0.113.2", "IP6Addr": "2001:db8::2"}
				],
				"RegistrantContact": {
					"Name": "Alice Example",
					"Company": "Example Ltd",
					"Email": "alice@example.co.nz",
					"PostalAddress": {"Country": "NZ"}
				},
				"TechnicalContact": {
					"Name": "Bob Example",
					"Company": "",
					"Email": "bob@example.co.nz",
					"PostalAddress": {"Country": "NZ"}
				},
				"AdminContact": {
					"Name": "Carol Example",
					"Company": "",
					"Email": "carol@example.co.nz",
					"PostalAddress": {"Country": "NZ"}
				}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).Whois(context.Background(), WhoisOptions{Domain: "example.co.nz"})
	if err != nil {
		t.Fatalf("Whois: %v", err)
	}

	if !got.Status {
		t.Errorf("Status = false, want true")
	}
	if got.SourceIP != "203.0.113.10" {
		t.Errorf("SourceIP = %q, want 203.0.113.10", got.SourceIP)
	}
	if got.Return.Domain != "example.co.nz" {
		t.Errorf("Return.Domain = %q, want example.co.nz", got.Return.Domain)
	}
	if got.Return.State != "Active" {
		t.Errorf("Return.State = %q, want Active", got.Return.State)
	}
	if len(got.Return.NameServers) != 2 {
		t.Fatalf("len(NameServers) = %d, want 2", len(got.Return.NameServers))
	}
	if got.Return.NameServers[0].FQDN != "ns1.sitehost.nz" {
		t.Errorf("NameServers[0].FQDN = %q, want ns1.sitehost.nz", got.Return.NameServers[0].FQDN)
	}
	if got.Return.Registrant.Name != "Alice Example" {
		t.Errorf("Registrant.Name = %q, want Alice Example", got.Return.Registrant.Name)
	}
	if got.Return.Registrant.Company != "Example Ltd" {
		t.Errorf("Registrant.Company = %q, want Example Ltd", got.Return.Registrant.Company)
	}
	if got.Return.Registrant.Email != "alice@example.co.nz" {
		t.Errorf("Registrant.Email = %q, want alice@example.co.nz", got.Return.Registrant.Email)
	}
	if got := got.Return.Registrant.PostalAddress["Country"]; got != "NZ" {
		t.Errorf("Registrant.PostalAddress[Country] = %q, want NZ", got)
	}
}

func TestWhois_DomainRequired(t *testing.T) {
	t.Parallel()
	c, err := api.New("k", "1")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).Whois(context.Background(), WhoisOptions{})
	if err == nil {
		t.Fatal("Whois: expected error for empty Domain, got nil")
	}
	if !strings.Contains(err.Error(), "Domain is required") {
		t.Errorf("error = %q, want it to contain 'Domain is required'", err.Error())
	}
}
