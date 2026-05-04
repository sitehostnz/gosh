package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// NewUDAIOptions requests a fresh UDAI (transfer authorisation
// code) for a domain via "srs/new_udai.json".
//
// The UDAI is sent by the registry to the registrant's email on
// record. This endpoint does not return the code in its response;
// it triggers the email and returns a {status, msg} envelope.
//
// Use cases:
//
//   - Transferring a .nz domain *out* to another registrar — the
//     gaining registrar needs the UDAI from the registrant.
//   - Refreshing a stale or compromised UDAI.
//
// Behaviour is TLD-specific: .nz uses UDAI as its sole transfer-
// authorisation mechanism (no EPP-style transfer locks); gTLDs
// (.com, .net) use a different "auth code" concept that may or
// may not be exposed via this endpoint depending on registrar
// configuration.
type NewUDAIOptions struct {
	Domain string `url:"domain"`
}

// NewUDAI generates a fresh UDAI (auth code) for a domain via
// "srs/new_udai.json". The code is delivered by email to the
// registrant; it is not returned in the response.
func (s *Client) NewUDAI(ctx context.Context, opt NewUDAIOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.NewUDAI: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/new_udai.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// ValidateUDAIOptions checks whether a UDAI (auth code) is valid
// for a domain via "srs/validate_udai.json" (GET). Both fields are
// required.
//
// Used by the *gaining* registrar before submitting a transfer:
// validate the code first to catch typos or expired codes before
// kicking off the registry transfer process.
//
// The public docs do not list a parameter table; the inferred
// shape is domain=...&udai=.... Validate live before relying on
// the response shape.
type ValidateUDAIOptions struct {
	Domain string `url:"domain"`
	UDAI   string `url:"udai"`
}

// ValidateUDAI checks a UDAI code via "srs/validate_udai.json".
// Returns the bare {status, msg} envelope; status=true means the
// code is valid for the domain, false means rejected.
func (s *Client) ValidateUDAI(ctx context.Context, opt ValidateUDAIOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.ValidateUDAI: Domain is required")
	}
	if opt.UDAI == "" {
		return response, fmt.Errorf("srs.ValidateUDAI: UDAI is required")
	}
	path, err := net.AddOptions("srs/validate_udai.json", opt)
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
