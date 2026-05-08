package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// PrivacyOptions describes a privacy-protection toggle. Reason is
// required by the API for both enable and disable; it's recorded
// at the registrar but is not user-visible.
type PrivacyOptions struct {
	Domain string `url:"domain"`
	Reason string `url:"reason"`
}

// EnablePrivacyProtection enables WHOIS privacy on a domain via
// "srs/enable_privacy_protection.json". Reason is required by the
// API even though the field has no externally-visible effect.
//
// Verify the new state via srs.GetDomain(domain).Private
// (returned as "1" / "0" string).
func (s *Client) EnablePrivacyProtection(ctx context.Context, opt PrivacyOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.EnablePrivacyProtection: Domain is required")
	}
	if opt.Reason == "" {
		return response, fmt.Errorf("srs.EnablePrivacyProtection: Reason is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/enable_privacy_protection.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// DisablePrivacyProtection turns WHOIS privacy off via
// "srs/disable_privacy_protection.json". Same Reason-required
// shape as EnablePrivacyProtection.
func (s *Client) DisablePrivacyProtection(ctx context.Context, opt PrivacyOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.DisablePrivacyProtection: Domain is required")
	}
	if opt.Reason == "" {
		return response, fmt.Errorf("srs.DisablePrivacyProtection: Reason is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/disable_privacy_protection.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
