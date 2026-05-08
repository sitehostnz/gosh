package srs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestCreateContact_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/create_contact.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("name") != "Test" || v.Get("email") != "t@example.com" {
			t.Errorf("body = %v", v)
		}
		if v.Get("country") != "NZ" || v.Get("params[PostCode]") != "1010" {
			t.Errorf("missing required: %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		// Phone fields encode as params[Phone][Country] etc.
		if v.Get("params[Phone][Country]") != "64" {
			t.Errorf("Phone[Country] = %q", v.Get("params[Phone][Country]"))
		}
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"ContactID":"233801"}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).CreateContact(context.Background(), CreateContactOptions{
		Name: "Test", Email: "t@example.com",
		PostalAddress: "1 Test St", PostalAddress2: "",
		City: "Auckland", Country: "NZ", PostCode: "1010",
		PhoneCountry: "64", PhoneArea: "9", PhoneLocal: "9742182",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if got.Return.ContactID != "233801" {
		t.Errorf("ContactID = %q", got.Return.ContactID)
	}
}

func TestCreateContact_RequiresNameAndEmail(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).CreateContact(context.Background(), CreateContactOptions{Email: "x"}); err == nil {
		t.Fatal("expected error for missing Name")
	}
	if _, err := New(c).CreateContact(context.Background(), CreateContactOptions{Name: "x"}); err == nil {
		t.Fatal("expected error for missing Email")
	}
}

func TestUpdateContact_OmitsEmptyFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("contact_id") != "100" {
			t.Errorf("contact_id = %q", v.Get("contact_id"))
		}
		if v.Get("params[Email]") != "new@example.com" {
			t.Errorf("Email = %q", v.Get("params[Email]"))
		}
		// City wasn't set → must be omitted
		if _, ok := v["params[City]"]; ok {
			t.Errorf("params[City] should be omitted")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateContact(context.Background(), UpdateContactOptions{
		ContactID: 100, Email: "new@example.com",
	}); err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}
}

func TestDeleteContact_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/delete_contact.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("contact_id") != "12" {
			t.Errorf("contact_id = %q", v.Get("contact_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).DeleteContact(context.Background(), DeleteContactOptions{ContactID: 12}); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}
}

func TestUpdateDomainContacts_RequiresThreeRoles(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	// Missing TechnicalContactID
	if _, err := New(c).UpdateDomainContacts(context.Background(), UpdateDomainContactsOptions{
		Domain: "example.com", RegistrantContactID: 1, AdminContactID: 2,
	}); err == nil {
		t.Fatal("expected error for missing TechnicalContactID")
	}
}

func TestUpdateDomainContacts_BillingOptional(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("registrant_contact_id") != "10" {
			t.Errorf("registrant = %q", v.Get("registrant_contact_id"))
		}
		// Billing not set → should be omitted
		if _, ok := v["billing_contact_id"]; ok {
			t.Errorf("billing should be omitted when zero")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateDomainContacts(context.Background(), UpdateDomainContactsOptions{
		Domain: "example.com", RegistrantContactID: 10,
		AdminContactID: 11, TechnicalContactID: 12,
	}); err != nil {
		t.Fatalf("UpdateDomainContacts: %v", err)
	}
}

func TestUpdateCompanyInfo_OnlySendsSetFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("params[CompanyName]") != "TestCo" {
			t.Errorf("CompanyName = %q", v.Get("params[CompanyName]"))
		}
		if _, ok := v["params[CompanyPhone]"]; ok {
			t.Errorf("CompanyPhone should be omitted when unset")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateCompanyInfo(context.Background(), UpdateCompanyInfoOptions{
		CompanyName: "TestCo",
	}); err != nil {
		t.Fatalf("UpdateCompanyInfo: %v", err)
	}
}
