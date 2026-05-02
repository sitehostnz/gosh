package srs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListDomains_Success(t *testing.T) {
	const (
		apiKey   = "test-key"
		clientID = "1234"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/srs/list_domains.json" {
			t.Errorf("path = %q, want /srs/list_domains.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": {
				"current_items": 2,
				"current_page": 1,
				"total_pages": 1,
				"total_items": 2,
				"data": [
					{
						"domain_id": "100", "domain": "example.co.nz", "state": "Active", "api": "IRS",
						"client_id": "1234", "client_name": "Test",
						"locked": "0", "private": "0", "pending": "0", "premium": "0",
						"registrant_name": "Alice", "reg_id": "10", "reg_name": "Alice",
						"adm_id": "11", "adm_name": "Alice", "tec_id": "12", "tec_name": "Alice",
						"autorenew_term": "12", "autorenew_days_remaining": "60",
						"dateregistered":  "2025-01-01 00:00:00", "datemodified": "2025-01-02 00:00:00",
						"datebilleduntil": "2026-01-01 00:00:00",
						"datecancelled":   "0000-00-00 00:00:00", "datelocked": "0000-00-00 00:00:00"
					},
					{
						"domain_id": "101", "domain": "another.nz", "state": "Active", "api": "IRS",
						"client_id": "1234", "client_name": "Test",
						"locked": "0", "private": "0", "pending": "0", "premium": "0",
						"registrant_name": "Bob", "reg_id": "20", "reg_name": "Bob",
						"adm_id": "21", "adm_name": "Bob", "tec_id": "22", "tec_name": "Bob",
						"autorenew_term": "12", "autorenew_days_remaining": "30",
						"dateregistered":  "2024-06-01 00:00:00", "datemodified": "2024-06-02 00:00:00",
						"datebilleduntil": "2026-06-01 00:00:00",
						"datecancelled":   "0000-00-00 00:00:00", "datelocked": "0000-00-00 00:00:00"
					}
				]
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New(apiKey, clientID, api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListDomains(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDomains: %v", err)
	}

	if !got.Status {
		t.Errorf("Status = false, want true")
	}
	if len(got.Return.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(got.Return.Data))
	}
	if got.Return.Data[0].Domain != "example.co.nz" {
		t.Errorf("Data[0].Domain = %q, want example.co.nz", got.Return.Data[0].Domain)
	}
	if got.Return.Data[0].ID != "100" {
		t.Errorf("Data[0].ID = %q, want 100", got.Return.Data[0].ID)
	}
	if got.Return.Data[0].ClientID != "1234" {
		t.Errorf("Data[0].ClientID = %q, want 1234", got.Return.Data[0].ClientID)
	}
	if got.Return.Data[0].RegistrantName != "Alice" {
		t.Errorf("Data[0].RegistrantName = %q, want Alice", got.Return.Data[0].RegistrantName)
	}
	if got.Return.Data[0].AutoRenewTerm != "12" {
		t.Errorf("Data[0].AutoRenewTerm = %q, want 12", got.Return.Data[0].AutoRenewTerm)
	}
	if got.Return.Data[0].Locked != "0" {
		t.Errorf("Data[0].Locked = %q, want 0", got.Return.Data[0].Locked)
	}
}

func TestListDomains_FilterParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("filters[sort_by]"); got != "domain" {
			t.Errorf("filters[sort_by] = %q, want domain", got)
		}
		if got := q.Get("filters[sort_dir]"); got != "asc" {
			t.Errorf("filters[sort_dir] = %q, want asc", got)
		}
		if got := q.Get("filters[page_size]"); got != "50" {
			t.Errorf("filters[page_size] = %q, want 50", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"data":[]}}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).ListDomains(context.Background(), &ListDomainsOptions{
		SortBy:   "domain",
		SortDir:  "asc",
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("ListDomains: %v", err)
	}
}
