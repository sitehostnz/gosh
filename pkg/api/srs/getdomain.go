package srs

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetDomain returns the full record for a domain via
// "srs/get_domain.json". Domain is required.
func (s *Client) GetDomain(ctx context.Context, opt DomainOptions) (response GetDomainResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.GetDomain: Domain is required")
	}
	path, err := net.AddOptions("srs/get_domain.json", opt)
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

// DomainAvailable reports whether a domain is registrable via
// "srs/domain_available.json". Domain is required.
func (s *Client) DomainAvailable(ctx context.Context, opt DomainAvailableOptions) (response BoolResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.DomainAvailable: Domain is required")
	}
	path, err := net.AddOptions("srs/domain_available.json", opt)
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

// DomainInsideGracePeriod reports whether a domain is in the
// post-cancellation grace period via
// "srs/domain_inside_grace_period.json". Domain is required.
func (s *Client) DomainInsideGracePeriod(ctx context.Context, opt DomainOptions) (response BoolResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.DomainInsideGracePeriod: Domain is required")
	}
	path, err := net.AddOptions("srs/domain_inside_grace_period.json", opt)
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

// GetDomainPrice returns pricing for a domain via
// "srs/get_domain_price.json". Domain is required.
func (s *Client) GetDomainPrice(ctx context.Context, opt DomainOptions) (response GetDomainPriceResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.GetDomainPrice: Domain is required")
	}
	path, err := net.AddOptions("srs/get_domain_price.json", opt)
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

// CanTransferDomain reports whether a domain can be transferred
// in via "srs/can_transfer_domain.json". Domain is required.
func (s *Client) CanTransferDomain(ctx context.Context, opt DomainOptions) (response CanTransferDomainResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.CanTransferDomain: Domain is required")
	}
	path, err := net.AddOptions("srs/can_transfer_domain.json", opt)
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

// ListNameServers returns the nameserver delegation for a
// domain via "srs/list_name_servers.json". Domain is required.
func (s *Client) ListNameServers(ctx context.Context, opt DomainOptions) (response ListNameServersResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.ListNameServers: Domain is required")
	}
	path, err := net.AddOptions("srs/list_name_servers.json", opt)
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
