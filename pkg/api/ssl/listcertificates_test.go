package ssl

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListCertificates_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ssl/list_certificates.json" {
			t.Errorf("path = %q, want /ssl/list_certificates.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": [
				{"cert_id": "1911", "common_name": "example.co.nz", "issue_date": "0000-00-00 00:00:00", "expiry_date": "0000-00-00 00:00:00"},
				{"cert_id": "1912", "common_name": "another.nz",    "issue_date": "2025-01-01 00:00:00", "expiry_date": "2026-01-01 00:00:00"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListCertificates(context.Background())
	if err != nil {
		t.Fatalf("ListCertificates: %v", err)
	}

	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	if got.Return[0].CertID != "1911" {
		t.Errorf("Return[0].CertID = %q, want 1911", got.Return[0].CertID)
	}
	if got.Return[1].CommonName != "another.nz" {
		t.Errorf("Return[1].CommonName = %q, want another.nz", got.Return[1].CommonName)
	}
}
