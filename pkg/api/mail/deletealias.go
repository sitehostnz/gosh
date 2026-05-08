package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// DeleteAlias removes an alias mapping via
// "mail/delete_alias.json". ServerName, Source, and Destination
// are all required — the API rejects calls with only Source.
// Synchronous; returns models.APIResponse only.
func (s *Client) DeleteAlias(ctx context.Context, opt DeleteAliasOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.DeleteAlias: ServerName is required (or set via NewForServer)")
	}
	if opt.Source == "" {
		return response, fmt.Errorf("mail.DeleteAlias: Source is required")
	}
	if opt.Destination == "" {
		return response, fmt.Errorf("mail.DeleteAlias: Destination is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/delete_alias.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
