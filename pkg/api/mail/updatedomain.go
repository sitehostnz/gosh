package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// UpdateDomain updates a mail domain via
// "mail/update_domain.json". ServerName and Domain are required;
// Params.Catchall and Params.State (string "0"/"1") are
// optional. Synchronous; returns models.APIResponse only.
func (s *Client) UpdateDomain(ctx context.Context, opt UpdateDomainOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.UpdateDomain: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.UpdateDomain: Domain is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/update_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
