package ssl

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetCertificate retrieves the full record for a single certificate
// (including the PEM-encoded CSR, chain, and CRT when issued) via
// "ssl/get_certificate.json". The CertID field of opt is required.
func (s *Client) GetCertificate(ctx context.Context, opt CertificateOptions) (response GetCertificateResponse, err error) {
	if opt.CertID == "" {
		return response, fmt.Errorf("ssl.GetCertificate: CertID is required")
	}

	u := "ssl/get_certificate.json"

	path, err := net.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
