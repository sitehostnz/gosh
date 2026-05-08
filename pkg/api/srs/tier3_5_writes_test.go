package srs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const testTemplate = "renewal"

// ---------------- Tier 3: domain mutators ----------------

func TestAddNameServers_EncodesArray(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/add_name_servers.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("domain") != testComDomain ||
			v.Get("nameservers[0][name]") != "ns1.example.com" ||
			v.Get("nameservers[1][name]") != "ns2.example.com" ||
			v.Get("nameservers[0][ipv4addr]") != "192.0.2.1" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).AddNameServers(context.Background(), AddNameServersOptions{
		Domain: testComDomain,
		NameServers: []NameServerEntry{
			{Name: "ns1.example.com", IPv4Addr: "192.0.2.1"},
			{Name: "ns2.example.com"},
		},
	}); err != nil {
		t.Fatalf("AddNameServers: %v", err)
	}
}

func TestAddNameServers_RequiresAtLeastOne(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).AddNameServers(context.Background(), AddNameServersOptions{Domain: testComDomain}); err == nil {
		t.Fatal("expected error for empty NameServers")
	}
}

func TestRenewDomain_RequiresTerm(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).RenewDomain(context.Background(), RenewDomainOptions{Domain: testComDomain}); err == nil {
		t.Fatal("expected error for Term=0")
	}
}

func TestRenewDomain_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("term") != "12" || v.Get("options[privacy]") != "1" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).RenewDomain(context.Background(), RenewDomainOptions{
		Domain: testComDomain, Term: 12, Privacy: "1",
	}); err != nil {
		t.Fatalf("RenewDomain: %v", err)
	}
}

func TestUpdateDomain_RequiresDomain(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).UpdateDomain(context.Background(), UpdateDomainOptions{}); err == nil {
		t.Fatal("expected error for missing Domain")
	}
}

func TestTLDsAvailable_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q", r.Method)
		}
		if r.URL.Query().Get("domain") != testComDomain {
			t.Errorf("query = %v", r.URL.Query())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"domain":"example.com","available":true}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).TLDsAvailable(context.Background(), TLDsAvailableOptions{Domain: testComDomain})
	if err != nil {
		t.Fatalf("TLDsAvailable: %v", err)
	}
	if !got.Return.Available {
		t.Errorf("Available = %v", got.Return.Available)
	}
}

// ---------------- Tier 4: UDAI / transfer ----------------

func TestNewUDAI_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/new_udai.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("domain") != testComDomain {
			t.Errorf("domain = %q", v.Get("domain"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).NewUDAI(context.Background(), NewUDAIOptions{Domain: testComDomain}); err != nil {
		t.Fatalf("NewUDAI: %v", err)
	}
}

func TestValidateUDAI_RequiresBoth(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).ValidateUDAI(context.Background(), ValidateUDAIOptions{Domain: "x"}); err == nil {
		t.Fatal("expected error for missing UDAI")
	}
	if _, err := New(c).ValidateUDAI(context.Background(), ValidateUDAIOptions{UDAI: "x"}); err == nil {
		t.Fatal("expected error for missing Domain")
	}
}

func TestValidateUDAI_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("domain") != testComDomain || q.Get("udai") != "ABC123" {
			t.Errorf("query = %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).ValidateUDAI(context.Background(), ValidateUDAIOptions{
		Domain: testComDomain, UDAI: "ABC123",
	}); err != nil {
		t.Fatalf("ValidateUDAI: %v", err)
	}
}

func TestTransferDomain_EncodesAllParams(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("domain") != testComDomain || v.Get("udai") != "ABC" {
			t.Errorf("body = %v", v)
		}
		if v.Get("params[registrant_contact_id]") != "10" {
			t.Errorf("registrant = %q", v.Get("params[registrant_contact_id]"))
		}
		if v.Get("params[term]") != "12" {
			t.Errorf("term = %q", v.Get("params[term]"))
		}
		if v.Get("params[nameservers][0][name]") != "ns1.example.com" {
			t.Errorf("ns0 = %q", v.Get("params[nameservers][0][name]"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).TransferDomain(context.Background(), TransferDomainOptions{
		Domain: testComDomain, UDAI: "ABC",
		RegistrantContactID: 10, AdminContactID: 11, TechnicalContactID: 12,
		Term:        12,
		NameServers: []NameServerEntry{{Name: "ns1.example.com"}},
	}); err != nil {
		t.Fatalf("TransferDomain: %v", err)
	}
}

func TestVerifyEmailToken_RequiresToken(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).VerifyEmailToken(context.Background(), VerifyEmailTokenOptions{}); err == nil {
		t.Fatal("expected error for missing Token")
	}
}

// ---------------- Tier 5: email templates ----------------

func TestListEmailTemplates_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/srs/list_email_templates.json" {
			t.Errorf("method/path = %q %q", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[{"name":"renewal","subject":"S","template":"T","type":"AutoRenewReminder"}]}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListEmailTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListEmailTemplates: %v", err)
	}
	if len(got.Return) != 1 || got.Return[0].Name != testTemplate {
		t.Errorf("Return = %+v", got.Return)
	}
}

// TestGetEmailTemplate_Unsupported asserts the wrapper short-circuits
// with ErrEmailTemplateUnsupported (no API round-trip) so callers
// detect the broken-upstream state via errors.Is. When the API's
// expected input shape is clarified, switch this test back to the
// round-trip form.
func TestGetEmailTemplate_Unsupported(t *testing.T) {
	t.Parallel()
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	_, err := New(c).GetEmailTemplate(context.Background(), GetEmailTemplateOptions{Template: testTemplate})
	if !errors.Is(err, ErrEmailTemplateUnsupported) {
		t.Fatalf("err = %v, want errors.Is(_, ErrEmailTemplateUnsupported)", err)
	}
	if called {
		t.Errorf("expected no HTTP call; the wrapper should short-circuit")
	}
}

func TestUpdateEmailTemplate_OmitsEmpty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("template") != "renewal" {
			t.Errorf("template = %q", v.Get("template"))
		}
		if v.Get("params[EmailSubject]") != "New" {
			t.Errorf("subject = %q", v.Get("params[EmailSubject]"))
		}
		if _, ok := v["params[EmailTemplate]"]; ok {
			t.Errorf("EmailTemplate should be omitted when empty")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK"}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateEmailTemplate(context.Background(), UpdateEmailTemplateOptions{
		Template: "renewal", EmailSubject: "New",
	}); err != nil {
		t.Fatalf("UpdateEmailTemplate: %v", err)
	}
}

func TestUpdateEmailTemplate_RequiresTemplate(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).UpdateEmailTemplate(context.Background(), UpdateEmailTemplateOptions{}); err == nil {
		t.Fatal("expected error for missing Template")
	}
}
