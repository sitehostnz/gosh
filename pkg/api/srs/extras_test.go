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

// mockSRS spins up an httptest server and a client pointed at it.
func mockSRS(t *testing.T, h http.HandlerFunc) (*api.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		srv.Close()
		t.Fatalf("api.New: %v", err)
	}
	return c, srv.Close
}

func TestGetDomain_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/get_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("domain"); got != "example.nz" {
			t.Errorf("domain = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful.",
			"return": {
				"ClientID": 1234, "Domain": "example.nz", "State": "Active",
				"RState": 0, "API": "IRS", "RegistrantName": "Alice",
				"DateRegistered": "2015-12-16 11:37:19",
				"DateBilledUntil": "2034-09-16 11:37:19",
				"autorenew_term": 12, "autorenew_days_remaining": 60,
				"RegistrantContactID": 9001, "AdminContactID": 9002,
				"TechnicalContactID": 9003, "BillingContactID": 9004,
				"Locked": false, "Private": false, "Pending": false,
				"Premium": false
			}
		}`)
	})
	defer done()

	got, err := New(c).GetDomain(context.Background(), DomainOptions{Domain: "example.nz"})
	if err != nil {
		t.Fatalf("GetDomain: %v", err)
	}
	if got.Return.Domain != "example.nz" {
		t.Errorf("Domain = %q", got.Return.Domain)
	}
	if got.Return.AutorenewTerm != 12 {
		t.Errorf("AutorenewTerm = %d", got.Return.AutorenewTerm)
	}
	if got.Return.Locked {
		t.Errorf("Locked = true (real bool, not string)")
	}
}

func TestDomainAvailable_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Available","return":true}`)
	})
	defer done()
	got, err := New(c).DomainAvailable(context.Background(), DomainAvailableOptions{Domain: "fresh.nz"})
	if err != nil {
		t.Fatalf("DomainAvailable: %v", err)
	}
	if !got.Return {
		t.Errorf("Return = false, want true")
	}
}

func TestDomainInsideGracePeriod_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Not in grace period","return":false}`)
	})
	defer done()
	got, err := New(c).DomainInsideGracePeriod(context.Background(), DomainOptions{Domain: "example.nz"})
	if err != nil {
		t.Fatalf("DomainInsideGracePeriod: %v", err)
	}
	if got.Return {
		t.Errorf("Return = true, want false")
	}
}

func TestGetDomainPrice_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"DomainPrice":39.5,"total_price":39.5,"tiered_price":"39.50","base_price":"39.50","premium":false,"base_privacy_price":"0.00","tiered_privacy_price":"0.00"}}`)
	})
	defer done()
	got, err := New(c).GetDomainPrice(context.Background(), DomainOptions{Domain: "example.nz"})
	if err != nil {
		t.Fatalf("GetDomainPrice: %v", err)
	}
	if got.Return.DomainPrice != 39.5 {
		t.Errorf("DomainPrice = %v", got.Return.DomainPrice)
	}
}

func TestCanTransferDomain_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"domain":"example.nz","can_transfer":true,"reason":""}}`)
	})
	defer done()
	got, err := New(c).CanTransferDomain(context.Background(), DomainOptions{Domain: "example.nz"})
	if err != nil {
		t.Fatalf("CanTransferDomain: %v", err)
	}
	if !got.Return.CanTransfer {
		t.Errorf("CanTransfer = false")
	}
}

func TestListNameServers_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[
			{"name":"ns1.sitehost.co.nz","ipv4addr":"","ipv6addr":""},
			{"name":"ns2.sitehost.co.nz","ipv4addr":"","ipv6addr":""}
		]}`)
	})
	defer done()
	got, err := New(c).ListNameServers(context.Background(), DomainOptions{Domain: "example.nz"})
	if err != nil {
		t.Fatalf("ListNameServers: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
}

func TestListContacts_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"total_items":1,"current_items":1,"current_page":1,"total_pages":1,
			"status":true,"msg":"OK",
			"return":[{"contact_id":"9001","name":"Alice","registrant_name":"Alice","email":"a@example.co.nz","phone_cntry":"64","phone_area":"09","phone_local":"1234567","phone_extension":"","domain_count":1}]
		}`)
	})
	defer done()
	got, err := New(c).ListContacts(context.Background())
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].DomainCount != 1 {
		t.Errorf("DomainCount = %d", got.Return[0].DomainCount)
	}
}

func TestGetContact_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("contact_id"); got != "9001" {
			t.Errorf("contact_id = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"OK",
			"return":{"ContactID":9001,"ClientID":"1234","Name":"Alice","RegistrantName":"Alice","Country":"NZ"}
		}`)
	})
	defer done()
	got, err := New(c).GetContact(context.Background(), ContactOptions{ContactID: "9001"})
	if err != nil {
		t.Fatalf("GetContact: %v", err)
	}
	if got.Return.Country != "NZ" {
		t.Errorf("Country = %q", got.Return.Country)
	}
}

func TestSearchContacts_FilterRequired(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).SearchContacts(context.Background(), SearchContactsOptions{})
	if err == nil || !strings.Contains(err.Error(), "at least one of") {
		t.Errorf("expected filter required, got: %v", err)
	}
}

func TestListValidTLDs_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[
			{"tld":"co.nz","term":"12","type":"NZRS","can_transfer":"1","can_protect":"1","date_added":"","date_updated":""}
		]}`)
	})
	defer done()
	got, err := New(c).ListValidTLDs(context.Background())
	if err != nil {
		t.Fatalf("ListValidTLDs: %v", err)
	}
	if got.Return[0].TLD != "co.nz" {
		t.Errorf("TLD = %q", got.Return[0].TLD)
	}
}

func TestGetPricingTiers_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[
			{"type":1,"type_name":"NZRS","count":"19","price":"39.50"}
		]}`)
	})
	defer done()
	got, err := New(c).GetPricingTiers(context.Background())
	if err != nil {
		t.Fatalf("GetPricingTiers: %v", err)
	}
	if got.Return[0].Price != "39.50" {
		t.Errorf("Price = %q", got.Return[0].Price)
	}
}

func TestGetCompanyInfo_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"client_id":1234,"companyname":"Test Co","companyurl":"https://test","companyrenewurl":"","companyemail":"","companyemailfrom":"","companyemailfromname":"","companysupportemail":"","companyphone":"","companyfax":"","send_renewed_email":"1","send_invoice":"1"}}`)
	})
	defer done()
	got, err := New(c).GetCompanyInfo(context.Background())
	if err != nil {
		t.Fatalf("GetCompanyInfo: %v", err)
	}
	if got.Return.CompanyName != "Test Co" {
		t.Errorf("CompanyName = %q", got.Return.CompanyName)
	}
}

func TestCreateDomain_RequiredFields(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).CreateDomain(context.Background(), CreateDomainOptions{Domain: "x.nz"})
	if err == nil || !strings.Contains(err.Error(), "contact IDs") {
		t.Errorf("expected contact IDs required, got: %v", err)
	}
}

func TestCancelDomain_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/cancel_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		if got := r.Form.Get("domain"); got != "x.nz" {
			t.Errorf("domain = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":29658613,"type":"scheduler"}}}`)
	})
	defer done()
	got, err := New(c).CancelDomain(context.Background(), DomainOptions{Domain: "x.nz"})
	if err != nil {
		t.Fatalf("CancelDomain: %v", err)
	}
	if got.Return.Job.ID != 29658613 {
		t.Errorf("Job.ID = %d", got.Return.Job.ID)
	}
}

func TestCreateDomain_Success(t *testing.T) {
	c, done := mockSRS(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/create_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		if got := r.Form.Get("registrant_contact"); got != "9001" {
			t.Errorf("registrant_contact = %q (note: not registrant_contact_id)", got)
		}
		if got := r.Form.Get("params[AdminContact]"); got != "9002" {
			t.Errorf("params[AdminContact] = %q (note: PascalCase nested)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":29658610,"type":"scheduler"}}}`)
	})
	defer done()
	got, err := New(c).CreateDomain(context.Background(), CreateDomainOptions{
		Domain:            "fresh.nz",
		Term:              12,
		RegistrantContact: 9001,
		AdminContact:      9002,
		TechnicalContact:  9003,
		BillingContact:    9004,
	})
	if err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}
	if got.Return.Job.ID != 29658610 {
		t.Errorf("Job.ID = %d", got.Return.Job.ID)
	}
}
