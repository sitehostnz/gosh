package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// DeleteAliasDomain removes an alias-domain mapping via
// "mail/delete_alias_domain.json". ServerName and AliasDomain
// are required. Synchronous; returns models.APIResponse only.
func (s *Client) DeleteAliasDomain(ctx context.Context, opt DeleteAliasDomainOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.DeleteAliasDomain: ServerName is required (or set via NewForServer)")
	}
	if opt.AliasDomain == "" {
		return response, fmt.Errorf("mail.DeleteAliasDomain: AliasDomain is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/delete_alias_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
