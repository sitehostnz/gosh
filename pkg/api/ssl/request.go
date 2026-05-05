package ssl

type (
	// CertificateOptions identifies a certificate by ID. Used by
	// GetCertificate and GetCSR.
	CertificateOptions struct {
		CertID string `url:"cert_id"`
	}
)
