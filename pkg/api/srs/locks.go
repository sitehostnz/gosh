package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// LockDomain places a transfer lock on a domain via
// "srs/lock_domain.json". Lock prevents transfer-out attempts at
// the registry until the domain is unlocked. Synchronous; the
// response is a bare {status, msg} envelope (no scheduler job).
//
// **TLD-policy gotcha (verified live):** **.nz domains reject this
// call** with "This domain cannot be locked." The .nz registry
// uses the UDAI (transfer authorisation code) model rather than
// EPP-style transfer locks, so transfer protection is achieved by
// withholding the UDAI rather than by locking. Other gTLDs (.com,
// .net, etc.) honour the lock as expected.
//
// Verify the new state via srs.GetDomain(domain).DateLocked
// (empty / zero-date when unlocked).
func (s *Client) LockDomain(ctx context.Context, opt DomainOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.LockDomain: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/lock_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UnlockDomain releases the transfer lock on a domain via
// "srs/unlock_domain.json". Same shape as LockDomain — bare
// {status, msg} envelope, idempotent.
//
// Same .nz caveat applies: .nz registry policy rejects the call
// (UDAI model, not lock-based transfer auth). See LockDomain.
func (s *Client) UnlockDomain(ctx context.Context, opt DomainOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.UnlockDomain: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/unlock_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
