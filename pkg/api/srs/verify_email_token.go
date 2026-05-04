package srs

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// VerifyEmailTokenOptions verifies a registrant-email confirmation
// token via "srs/verify_email_token.json" (GET).
//
// Background: when a contact's email changes (or a new domain is
// registered), some registries (notably ICANN-governed gTLDs)
// require the registrant to confirm ownership of the email by
// clicking a link. The link carries a token; this endpoint marks
// the token as confirmed.
//
// The public docs do not list parameters; the inferred shape is
// `?token=...`. Validate live before relying on the response shape.
//
// This endpoint is typically driven by the end-user clicking a
// link in the verification email rather than called from server
// code; it's wrapped here for completeness and for any automated
// re-verification flows.
type VerifyEmailTokenOptions struct {
	Token string `url:"token"`
}

// VerifyEmailToken confirms a registrant-email verification token
// via "srs/verify_email_token.json".
func (s *Client) VerifyEmailToken(ctx context.Context, opt VerifyEmailTokenOptions) (response models.APIResponse, err error) {
	if opt.Token == "" {
		return response, fmt.Errorf("srs.VerifyEmailToken: Token is required")
	}
	path, err := net.AddOptions("srs/verify_email_token.json", opt)
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
