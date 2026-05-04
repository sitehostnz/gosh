package srs

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// TLDsAvailableOptions checks domain availability via
// "srs/domain_tlds_available.json" (GET).
//
// Despite the plural-TLD endpoint name, live behaviour (May 2026)
// is single-domain: requires a fully-qualified `domain=<label>.<tld>`
// and returns `{domain, available}` — the same shape as
// `srs/domain_available.json`. A bare label without TLD is rejected
// with "Please specify a valid domain name." If a true multi-TLD
// fan-out parameter exists it isn't documented and we couldn't
// surface it via probing — see docs/api-issues/.
type TLDsAvailableOptions struct {
	Domain string `url:"domain"`
}

// TLDsAvailableReturn is the inner payload — same shape as
// srs.DomainAvailable.
type TLDsAvailableReturn struct {
	Domain    string `json:"domain"`
	Available bool   `json:"available"`
}

// TLDsAvailableResponse mirrors the live single-domain payload.
type TLDsAvailableResponse struct {
	Return TLDsAvailableReturn `json:"return"`
	models.APIResponse
}

// TLDsAvailable performs a multi-TLD availability check via
// "srs/domain_tlds_available.json" (GET).
func (s *Client) TLDsAvailable(ctx context.Context, opt TLDsAvailableOptions) (response TLDsAvailableResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.TLDsAvailable: Domain is required")
	}
	path, err := net.AddOptions("srs/domain_tlds_available.json", opt)
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
