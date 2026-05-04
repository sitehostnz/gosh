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

func TestGetCSR_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ssl/get_csr.json" {
			t.Errorf("path = %q, want /ssl/get_csr.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("cert_id"); got != "1911" {
			t.Errorf("cert_id = %q, want 1911", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful.",
			"return": {
				"csr": {
					"csr_details": {
						"countryName": "NZ",
						"stateOrProvinceName": "Auckland",
						"localityName": "North Island",
						"organizationName": "Example Ltd",
						"organizationalUnitName": "Web",
						"commonName": "example.co.nz",
						"emailAddress": "admin@example.co.nz"
					},
					"csr": "-----BEGIN CERTIFICATE REQUEST-----\nMIIC...\n-----END CERTIFICATE REQUEST-----\n"
				}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetCSR(context.Background(), CertificateOptions{CertID: "1911"})
	if err != nil {
		t.Fatalf("GetCSR: %v", err)
	}

	d := got.Return.CSR.Details
	if d.CountryName != "NZ" {
		t.Errorf("Details.CountryName = %q, want NZ", d.CountryName)
	}
	if d.CommonName != "example.co.nz" {
		t.Errorf("Details.CommonName = %q, want example.co.nz", d.CommonName)
	}
	if d.OrganizationName != "Example Ltd" {
		t.Errorf("Details.OrganizationName = %q, want Example Ltd", d.OrganizationName)
	}
	if !strings.HasPrefix(got.Return.CSR.Raw, "-----BEGIN CERTIFICATE REQUEST-----") {
		t.Errorf("Return.CSR.Raw does not look like PEM")
	}
}

func TestGetCSR_CertIDRequired(t *testing.T) {
	t.Parallel()
	c, err := api.New("k", "1")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	_, err = New(c).GetCSR(context.Background(), CertificateOptions{})
	if err == nil {
		t.Fatal("expected error for empty CertID, got nil")
	}
	if !strings.Contains(err.Error(), "CertID is required") {
		t.Errorf("error = %q, want it to contain 'CertID is required'", err.Error())
	}
}
