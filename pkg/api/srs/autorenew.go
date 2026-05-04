package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// UpdateAutoRenewOptions configures the auto-renew schedule for a
// domain. All four fields are required by the API.
//
// Term is the renewal-period length in months (12 = 1 year, etc.).
// Set Term to 0 to disable auto-renew (verify by reading
// srs.GetDomain(domain).AutorenewTerm afterwards — value 0 means
// disabled).
//
// DaysRemaining controls how many days before expiry the renewal
// fires (e.g. 30 = renew 30 days before billed-until date).
type UpdateAutoRenewOptions struct {
	Domain        string `url:"domain"`
	Term          int    `url:"term"`
	DaysRemaining int    `url:"days_remaining"`
}

// UpdateAutoRenew configures the auto-renew schedule for a domain
// via "srs/update_auto_renew.json".
//
// Verify the new schedule via srs.GetDomain(domain).AutorenewTerm
// + AutorenewDaysRemaining.
func (s *Client) UpdateAutoRenew(ctx context.Context, opt UpdateAutoRenewOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.UpdateAutoRenew: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_auto_renew.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
