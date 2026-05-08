package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// DeleteForward removes a forwarder mapping via
// "mail/delete_forward.json". ServerName, Source, and
// Destination are all required — the API rejects calls with
// only Source. Synchronous; returns models.APIResponse only.
func (s *Client) DeleteForward(ctx context.Context, opt DeleteForwardOptions) (response models.APIResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.DeleteForward: ServerName is required (or set via NewForServer)")
	}
	if opt.Source == "" {
		return response, fmt.Errorf("mail.DeleteForward: Source is required")
	}
	if opt.Destination == "" {
		return response, fmt.Errorf("mail.DeleteForward: Destination is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/delete_forward.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
