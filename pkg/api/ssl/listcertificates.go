package ssl

import "context"

// ListCertificates retrieves the list of certificates registered for
// the authenticated client via "ssl/list_certificates.json". Each
// entry is a brief summary; use GetCertificate for the full record.
func (s *Client) ListCertificates(ctx context.Context) (response ListCertificatesResponse, err error) {
	u := "ssl/list_certificates.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
