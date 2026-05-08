package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// RenewDomainOptions extends a domain's registration term.
//
// **This call charges the SiteHost account** at the renewal price
// for Term months — same pricing matrix as srs.GetDomainPrice with
// type="renew". There is no dry-run mode. Read GetDomain.BilledUntil
// before and after to confirm the extension landed.
//
// Term is the renewal length in months (12 = 1 year, 24 = 2 years).
// Privacy is optional: 1 enables WHOIS privacy at renewal time, 0
// disables. Leave empty to keep the current setting.
type RenewDomainOptions struct {
	Domain  string `url:"domain"`
	Term    int    `url:"term"`
	Privacy string `url:"options[privacy],omitempty"`
}

// RenewDomain extends a domain's registration via
// "srs/renew_domain.json". Returns the bare {status, msg} envelope;
// the registry update is synchronous.
//
// **Cost warning:** this charges the account. Avoid in tests
// against billed test domains unless you've confirmed cost is
// acceptable.
func (s *Client) RenewDomain(ctx context.Context, opt RenewDomainOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.RenewDomain: Domain is required")
	}
	if opt.Term <= 0 {
		return response, fmt.Errorf("srs.RenewDomain: Term must be > 0 months")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/renew_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
