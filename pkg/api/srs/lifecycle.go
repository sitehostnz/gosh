package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// CreateDomain registers a new domain via
// "srs/create_domain.json". Domain and all four contact IDs
// (RegistrantContact, AdminContact, TechnicalContact,
// BillingContact) are required.
//
// **This is destructive.** Successful registration incurs a real
// registry charge. For .nz domains, registrations cancelled
// within 5 days of creation are not billed (per registry
// grace-period rules); other TLDs vary. The account must hold
// sufficient funds at registration time regardless.
//
// Returns a JobResponse with the scheduler job ID for tracking
// the asynchronous registration.
func (s *Client) CreateDomain(ctx context.Context, opt CreateDomainOptions) (response JobResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.CreateDomain: Domain is required")
	}
	if opt.RegistrantContact == 0 || opt.AdminContact == 0 ||
		opt.TechnicalContact == 0 || opt.BillingContact == 0 {
		return response, fmt.Errorf("srs.CreateDomain: all four contact IDs (RegistrantContact, AdminContact, TechnicalContact, BillingContact) are required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/create_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// CancelDomain cancels a domain registration via
// "srs/cancel_domain.json". Domain is required. Returns a
// JobResponse with the scheduler job ID for tracking the
// asynchronous cancellation.
//
// For .nz domains cancelled within 5 days of registration, no
// billing is incurred. Outside that window, billing rules apply
// per the registry.
func (s *Client) CancelDomain(ctx context.Context, opt DomainOptions) (response JobResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.CancelDomain: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/cancel_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
