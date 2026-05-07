package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// DeleteDomain removes a mail domain via
// "mail/delete_domain.json". ServerName and Domain are required.
// Synchronous; returns models.APIResponse only.
//
// The API refuses to delete a domain while aliases or forwards
// still exist on it — clean those up first.
func (s *Client) DeleteDomain(ctx context.Context, opt DeleteDomainOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.DeleteDomain: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.DeleteDomain: Domain is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/delete_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
