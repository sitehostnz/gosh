package ssl

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// CertificateSummary is the brief certificate entry returned by
	// list_certificates. The full certificate is available via
	// GetCertificate.
	CertificateSummary struct {
		CertID     string `json:"cert_id"`
		CommonName string `json:"common_name"`
		IssueDate  string `json:"issue_date"`
		ExpiryDate string `json:"expiry_date"`
	}

	// ListCertificatesResponse represents the response from
	// list_certificates.
	ListCertificatesResponse struct {
		Return []CertificateSummary `json:"return"`
		models.APIResponse
	}

	// Certificate is the full record for a certificate. Chain and CRT
	// are empty until the certificate is issued; CSR is the
	// PEM-encoded certificate signing request that was generated
	// when the cert was created.
	Certificate struct {
		CertID      string `json:"cert_id"`
		ClientID    string `json:"client_id"`
		CommonName  string `json:"common_name"`
		CSR         string `json:"csr"`
		Chain       string `json:"chain"`
		CRT         string `json:"crt"`
		IssueDate   string `json:"issue_date"`
		ExpiryDate  string `json:"expiry_date"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// GetCertificateResponse represents the response from
	// get_certificate.
	GetCertificateResponse struct {
		Return Certificate `json:"return"`
		models.APIResponse
	}

	// CSRDetails is the parsed subject information of a CSR.
	CSRDetails struct {
		CountryName            string `json:"countryName"`
		StateOrProvinceName    string `json:"stateOrProvinceName"`
		LocalityName           string `json:"localityName"`
		OrganizationName       string `json:"organizationName"`
		OrganizationalUnitName string `json:"organizationalUnitName"`
		CommonName             string `json:"commonName"`
		EmailAddress           string `json:"emailAddress"`
	}

	// CSR represents both the parsed details and the raw PEM-encoded
	// certificate signing request. The API wraps these inside the
	// outer "return.csr" object, so the parsed and raw views are
	// returned together.
	CSR struct {
		Details CSRDetails `json:"csr_details"`
		Raw     string     `json:"csr"`
	}

	// GetCSRResponse represents the response from get_csr.
	GetCSRResponse struct {
		Return struct {
			CSR CSR `json:"csr"`
		} `json:"return"`
		models.APIResponse
	}
)
