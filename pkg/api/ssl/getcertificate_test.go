package ssl

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGetCertificate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ssl/get_certificate.json" {
			t.Errorf("path = %q, want /ssl/get_certificate.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("cert_id"); got != "1911" {
			t.Errorf("cert_id = %q, want 1911", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": {
				"cert_id": "1911",
				"client_id": "1234",
				"common_name": "example.co.nz",
				"csr":   "-----BEGIN CERTIFICATE REQUEST-----\nMIIC...\n-----END CERTIFICATE REQUEST-----\n",
				"chain": "",
				"crt":   "",
				"issue_date":   "0000-00-00 00:00:00",
				"expiry_date":  "0000-00-00 00:00:00",
				"date_added":   "2026-05-01 12:00:00",
				"date_updated": "2026-05-01 12:00:00"
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetCertificate(context.Background(), CertificateOptions{CertID: "1911"})
	if err != nil {
		t.Fatalf("GetCertificate: %v", err)
	}

	if got.Return.CertID != "1911" {
		t.Errorf("Return.CertID = %q, want 1911", got.Return.CertID)
	}
	if got.Return.CommonName != "example.co.nz" {
		t.Errorf("Return.CommonName = %q, want example.co.nz", got.Return.CommonName)
	}
	if !strings.HasPrefix(got.Return.CSR, "-----BEGIN CERTIFICATE REQUEST-----") {
		t.Errorf("Return.CSR does not look like PEM: %q", got.Return.CSR[:min(60, len(got.Return.CSR))])
	}
	if got.Return.CRT != "" {
		t.Errorf("Return.CRT = %q, want empty (cert not yet issued)", got.Return.CRT)
	}
}

func TestGetCertificate_CertIDRequired(t *testing.T) {
	c, err := api.New("k", "1")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).GetCertificate(context.Background(), CertificateOptions{})
	if err == nil {
		t.Fatal("expected error for empty CertID, got nil")
	}
	if !strings.Contains(err.Error(), "CertID is required") {
		t.Errorf("error = %q, want it to contain 'CertID is required'", err.Error())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
