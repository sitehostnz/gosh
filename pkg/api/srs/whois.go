package srs

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Whois performs a whois lookup for the given domain via
// "srs/whois.json". The Domain field of opt is required.
//
// The endpoint is documented at https://docs.sitehost.nz/api/v1.5/?path=/srs.
// JSON keys in the response use PascalCase (DomainName, Status,
// RegisteredDate, ...) reflecting the underlying registry schema —
// see Whois and WhoisContact in whoisresponse.go.
func (s *Client) Whois(ctx context.Context, opt WhoisOptions) (response WhoisResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.Whois: Domain is required")
	}

	u := "srs/whois.json"

	path, err := net.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	httpReq, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, httpReq, &response); err != nil {
		return response, err
	}

	return response, nil
}
