package template

import (
	"context"
)

// List returns the DNS templates available to this client, via
// "dns/domain_templates/list_templates.json".
//
// # Shared templates appear alongside your own
//
// The listing includes SiteHost's shared templates as well as the
// account's. A shared template carries [DomainTemplate.ClientID] "0",
// so filter on that when you want only the account's own — and do not
// read DomainCount on a shared template as an account figure, since it
// is not scoped to the caller.
//
// The sentinel cannot be pinned by a recorded fixture: scrubbing
// collapses any digits-only string to "1", so a real "0" is
// unrepresentable in committed testdata. It is asserted in a
// hand-written test on the type instead.
//
// TemplateID "0" is a real, usable template rather than a null id,
// which is worth knowing before treating 0 as "unset".
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	req, err := s.client.NewRequest("GET", "dns/domain_templates/list_templates.json", "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
