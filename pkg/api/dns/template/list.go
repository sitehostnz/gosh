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
// account's. Filter on [DomainTemplate.ClientID] when you want only
// the latter, and do not read DomainCount on a shared template as an
// account figure — it is not scoped to the caller.
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
