package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// UpdateDomainOptions triggers a domain-record refresh via
// "srs/update_domain.json".
//
// The public docs only list client_id and domain as parameters;
// no params[*] block is documented. In practice the endpoint
// reconciles the local SiteHost record with the registry's view
// (whois, expiry, contact bindings, nameservers). It does not
// take new field values — for those use the role-specific
// endpoints (UpdateDomainContacts, AddNameServers, etc.).
//
// Returns the bare {status, msg} envelope. Re-read with
// srs.GetDomain afterwards to observe any changes.
type UpdateDomainOptions struct {
	Domain string `url:"domain"`
}

// UpdateDomain refreshes a domain record from the registry via
// "srs/update_domain.json".
func (s *Client) UpdateDomain(ctx context.Context, opt UpdateDomainOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.UpdateDomain: Domain is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
