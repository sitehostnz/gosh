package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// AddDomain adds a domain to the specified mail server via
// "mail/add_domain.json". ServerName and Domain are required.
// Synchronous; returns models.APIResponse only.
//
// Note: per the SiteHost docs, the domain must already be
// managed by the SiteHost Control Panel (i.e. exist as a DNS
// zone) before it can be added as a mail domain.
func (s *Client) AddDomain(ctx context.Context, opt AddDomainOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.AddDomain: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.AddDomain: Domain is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/add_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
