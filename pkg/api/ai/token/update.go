package token

import (
	"context"
	"errors"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update changes a token's label, capabilities or IP allowlist. Fields
// left unset on the request are not changed; see UpdateRequest.
//
// Changes to capabilities and the allowlist take effect on the AI
// platform immediately. A label is not sent to the AI platform, so a
// label-only change leaves the token's behaviour untouched.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	if request.KeyID == "" {
		return response, errors.New("token.Update: KeyID is required")
	}

	u := "ai/token/update.json"

	keys := []string{
		"key_id",
		"params[label]",
		"params[capabilities][]",
		"params[allowed_ips][]",
	}

	values := url.Values{}
	values.Add("key_id", request.KeyID)
	if request.Label != "" {
		values.Add("params[label]", request.Label)
	}
	for _, c := range request.Capabilities {
		values.Add("params[capabilities][]", c)
	}

	// Clearing the allowlist has to be sent as an array holding one
	// empty entry, params[allowed_ips][]= — the API drops blank entries,
	// leaving an empty list. The scalar params[allowed_ips]= looks like
	// the natural spelling and is wrong: it is accepted with
	// {"updated":true} and changes nothing, so a token meant to be
	// unrestricted stays locked to its old addresses.
	ips := request.AllowedIPs
	if ips != nil && len(ips) == 0 {
		ips = []string{""}
	}
	for _, ip := range ips {
		values.Add("params[allowed_ips][]", ip)
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
