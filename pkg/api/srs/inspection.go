package srs

import (
	"context"
)

// ListValidTLDs returns the list of TLDs the client may
// register via "srs/list_valid_tlds.json".
func (s *Client) ListValidTLDs(ctx context.Context) (response ListValidTLDsResponse, err error) {
	req, err := s.client.NewRequest("GET", "srs/list_valid_tlds.json", "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// GetPricingTiers returns the count→price tier map for domain
// registrations via "srs/get_pricing_tiers.json".
func (s *Client) GetPricingTiers(ctx context.Context) (response GetPricingTiersResponse, err error) {
	req, err := s.client.NewRequest("GET", "srs/get_pricing_tiers.json", "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// GetCompanyInfo returns the client's company profile (used in
// renewal emails and registry contacts) via
// "srs/get_company_info.json".
func (s *Client) GetCompanyInfo(ctx context.Context) (response GetCompanyInfoResponse, err error) {
	req, err := s.client.NewRequest("GET", "srs/get_company_info.json", "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
