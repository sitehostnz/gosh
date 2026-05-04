package srs

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

func TestLockDomain_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/lock_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("domain") != "example.com" {
			t.Errorf("domain = %q", v.Get("domain"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Domain locked."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).LockDomain(context.Background(), DomainOptions{Domain: "example.com"}); err != nil {
		t.Fatalf("LockDomain: %v", err)
	}
}

func TestLockDomain_RequiresDomain(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).LockDomain(context.Background(), DomainOptions{}); err == nil {
		t.Fatal("expected error for missing Domain")
	}
}

func TestUnlockDomain_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/unlock_domain.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Domain unlocked."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UnlockDomain(context.Background(), DomainOptions{Domain: "example.com"}); err != nil {
		t.Fatalf("UnlockDomain: %v", err)
	}
}

func TestEnablePrivacyProtection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/enable_privacy_protection.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("reason") != "test run" {
			t.Errorf("reason = %q", v.Get("reason"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Privacy enabled."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).EnablePrivacyProtection(context.Background(), PrivacyOptions{
		Domain: "example.com", Reason: "test run",
	}); err != nil {
		t.Fatalf("EnablePrivacyProtection: %v", err)
	}
}

func TestEnablePrivacyProtection_RequiresReason(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).EnablePrivacyProtection(context.Background(), PrivacyOptions{
		Domain: "example.com",
	}); err == nil {
		t.Fatal("expected error for missing Reason")
	}
}

func TestDisablePrivacyProtection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/disable_privacy_protection.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Privacy disabled."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).DisablePrivacyProtection(context.Background(), PrivacyOptions{
		Domain: "example.com", Reason: "test run",
	}); err != nil {
		t.Fatalf("DisablePrivacyProtection: %v", err)
	}
}

func TestUpdateAutoRenew_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/srs/update_auto_renew.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("term") != "12" || v.Get("days_remaining") != "30" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Auto-renew updated."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateAutoRenew(context.Background(), UpdateAutoRenewOptions{
		Domain: "example.com", Term: 12, DaysRemaining: 30,
	}); err != nil {
		t.Fatalf("UpdateAutoRenew: %v", err)
	}
}

func TestUpdateAutoRenew_DisableViaTermZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("term") != "0" {
			t.Errorf("term = %q (want 0 for disable)", v.Get("term"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Auto-renew disabled."}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateAutoRenew(context.Background(), UpdateAutoRenewOptions{
		Domain: "example.com", Term: 0, DaysRemaining: 30,
	}); err != nil {
		t.Fatalf("UpdateAutoRenew: %v", err)
	}
}
