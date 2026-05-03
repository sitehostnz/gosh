package mail

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// jobOK is a minimal canned response for write endpoints that
// return a JobResponse.
const jobOK = `{"status":true,"msg":"Successful","return":{"job":{"id":42,"type":"daemon"}}}`

// ackOK is a minimal canned response for write endpoints that
// return only models.APIResponse.
const ackOK = `{"status":true,"msg":"Successful."}`

// newMockClient spins up an httptest.Server with the given
// handler and returns a configured api.Client.
func newMockClient(t *testing.T, h http.HandlerFunc) (*api.Client, func()) {
	t.Helper()
	server := httptest.NewServer(h)
	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		server.Close()
		t.Fatalf("api.New: %v", err)
	}
	return c, server.Close
}

func formCheck(t *testing.T, r *http.Request, key, want string) {
	t.Helper()
	if got := r.Form.Get(key); got != want {
		t.Errorf("form[%q] = %q, want %q", key, got, want)
	}
}

// --- AddAccount -----------------------------------------------------

func TestAddAccount_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/add_account.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		formCheck(t, r, "email", "alice@example.co.nz")
		formCheck(t, r, "params[password]", "secret")
		formCheck(t, r, "params[label]", "Alice")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	})
	defer done()

	got, err := New(c).AddAccount(context.Background(), AddAccountOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Email:         "alice@example.co.nz",
		AccountParams: AccountParams{Password: "secret", Label: "Alice"},
	})
	if err != nil {
		t.Fatalf("AddAccount: %v", err)
	}
	if got.Return.Job.ID != 42 {
		t.Errorf("Job.ID = %d", got.Return.Job.ID)
	}
}

func TestAddAccount_RequiredFields(t *testing.T) {
	c, _ := api.New("k", "1")
	if _, err := New(c).AddAccount(context.Background(), AddAccountOptions{
		Email: "x", AccountParams: AccountParams{Password: "p"},
	}); err == nil || !strings.Contains(err.Error(), "ServerName is required") {
		t.Errorf("missing ServerName: %v", err)
	}
	if _, err := New(c).AddAccount(context.Background(), AddAccountOptions{
		ServerOptions: ServerOptions{ServerName: "s"}, AccountParams: AccountParams{Password: "p"},
	}); err == nil || !strings.Contains(err.Error(), "Email is required") {
		t.Errorf("missing Email: %v", err)
	}
	if _, err := New(c).AddAccount(context.Background(), AddAccountOptions{
		ServerOptions: ServerOptions{ServerName: "s"}, Email: "x",
	}); err == nil || !strings.Contains(err.Error(), "Password is required") {
		t.Errorf("missing Password: %v", err)
	}
}

// --- UpdateAccount --------------------------------------------------

func TestUpdateAccount_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/update_account.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	})
	defer done()

	if _, err := New(c).UpdateAccount(context.Background(), UpdateAccountOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Email:         "alice@example.co.nz",
		AccountParams: AccountParams{Label: "Alice (renamed)"},
	}); err != nil {
		t.Fatalf("UpdateAccount: %v", err)
	}
}

// --- DeleteAccount --------------------------------------------------

func TestDeleteAccount_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/delete_account.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		formCheck(t, r, "email", "alice@example.co.nz")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	})
	defer done()

	if _, err := New(c).DeleteAccount(context.Background(), DeleteAccountOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Email:         "alice@example.co.nz",
	}); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
}

// --- AddDomain ------------------------------------------------------

func TestAddDomain_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/add_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		formCheck(t, r, "domain", "example.co.nz")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	got, err := New(c).AddDomain(context.Background(), AddDomainOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
	})
	if err != nil {
		t.Fatalf("AddDomain: %v", err)
	}
	if !got.Status {
		t.Errorf("Status = false")
	}
}

// --- UpdateDomain ---------------------------------------------------

func TestUpdateDomain_Catchall(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		formCheck(t, r, "params[catchall]", "alice@example.co.nz")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).UpdateDomain(context.Background(), UpdateDomainOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
		DomainParams:  DomainParams{Catchall: "alice@example.co.nz"},
	}); err != nil {
		t.Fatalf("UpdateDomain: %v", err)
	}
}

// --- DeleteDomain ---------------------------------------------------

func TestDeleteDomain_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/delete_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).DeleteDomain(context.Background(), DeleteDomainOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Domain:        "example.co.nz",
	}); err != nil {
		t.Fatalf("DeleteDomain: %v", err)
	}
}

// --- AddAlias / DeleteAlias / AddForward / DeleteForward ------------

func TestAddAlias_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/add_alias.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		formCheck(t, r, "source", "info@example.co.nz")
		formCheck(t, r, "destination", "alice@example.co.nz")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	})
	defer done()

	if _, err := New(c).AddAlias(context.Background(), AddAliasOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "info@example.co.nz", Destination: "alice@example.co.nz",
	}); err != nil {
		t.Fatalf("AddAlias: %v", err)
	}
}

func TestDeleteAlias_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/delete_alias.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).DeleteAlias(context.Background(), DeleteAliasOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "info@example.co.nz", Destination: "alice@example.co.nz",
	}); err != nil {
		t.Fatalf("DeleteAlias: %v", err)
	}
}

func TestDeleteAlias_RequiresBoth(t *testing.T) {
	c, _ := api.New("k", "1")
	_, err := New(c).DeleteAlias(context.Background(), DeleteAliasOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "info@example.co.nz",
	})
	if err == nil || !strings.Contains(err.Error(), "Destination is required") {
		t.Errorf("expected Destination required, got: %v", err)
	}
}

func TestAddForward_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/add_forward.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	})
	defer done()

	if _, err := New(c).AddForward(context.Background(), AddForwardOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "external@example.co.nz", Destination: "remote@elsewhere.example",
	}); err != nil {
		t.Fatalf("AddForward: %v", err)
	}
}

func TestDeleteForward_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/delete_forward.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).DeleteForward(context.Background(), DeleteForwardOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		Source:        "external@example.co.nz", Destination: "remote@elsewhere.example",
	}); err != nil {
		t.Fatalf("DeleteForward: %v", err)
	}
}

// --- AddAliasDomain / DeleteAliasDomain -----------------------------

func TestAddAliasDomain_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/add_alias_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		formCheck(t, r, "alias_domain", "alias.example.co.nz")
		formCheck(t, r, "parent_domain", "example.co.nz")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).AddAliasDomain(context.Background(), AddAliasDomainOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		AliasDomain:   "alias.example.co.nz",
		ParentDomain:  "example.co.nz",
	}); err != nil {
		t.Fatalf("AddAliasDomain: %v", err)
	}
}

func TestDeleteAliasDomain_Success(t *testing.T) {
	c, done := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/delete_alias_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ackOK)
	})
	defer done()

	if _, err := New(c).DeleteAliasDomain(context.Background(), DeleteAliasDomainOptions{
		ServerOptions: ServerOptions{ServerName: "sth-mail-air"},
		AliasDomain:   "alias.example.co.nz",
	}); err != nil {
		t.Fatalf("DeleteAliasDomain: %v", err)
	}
}
