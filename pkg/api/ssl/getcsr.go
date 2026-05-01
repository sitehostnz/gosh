package ssl

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetCSR retrieves the certificate signing request (CSR) associated
// with a certificate via "ssl/get_csr.json". The response includes
// both the parsed subject details and the raw PEM-encoded CSR. The
// CertID field of opt is required.
func (s *Client) GetCSR(ctx context.Context, opt CertificateOptions) (response GetCSRResponse, err error) {
	if opt.CertID == "" {
		return response, fmt.Errorf("ssl.GetCSR: CertID is required")
	}

	u := "ssl/get_csr.json"

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
