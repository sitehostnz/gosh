package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// AddAccount creates a new mail account on the specified server
// via "mail/add_account.json". ServerName, Email, and
// Params.Password are required. Returns a JobResponse with the
// scheduler job ID.
func (s *Client) AddAccount(ctx context.Context, opt AddAccountOptions) (response JobResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.AddAccount: ServerName is required (or set via NewForServer)")
	}
	if opt.Email == "" {
		return response, fmt.Errorf("mail.AddAccount: Email is required")
	}
	if opt.Password == "" {
		return response, fmt.Errorf("mail.AddAccount: Password is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/add_account.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
