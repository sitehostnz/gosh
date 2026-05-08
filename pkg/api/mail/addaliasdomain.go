package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// AddAliasDomain adds an alias-domain mapping (one domain
// pointing at another for mail purposes) via
// "mail/add_alias_domain.json". ServerName, AliasDomain, and
// ParentDomain are required. Synchronous; returns
// models.APIResponse only.
func (s *Client) AddAliasDomain(ctx context.Context, opt AddAliasDomainOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.AddAliasDomain: ServerName is required (or set via NewForServer)")
	}
	if opt.AliasDomain == "" {
		return response, fmt.Errorf("mail.AddAliasDomain: AliasDomain is required")
	}
	if opt.ParentDomain == "" {
		return response, fmt.Errorf("mail.AddAliasDomain: ParentDomain is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/add_alias_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
