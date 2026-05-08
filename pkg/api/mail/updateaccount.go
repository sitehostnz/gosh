package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// UpdateAccount updates an existing mail account via
// "mail/update_account.json". ServerName and Email are required;
// any non-empty Params field is applied. Supplying no params is
// a no-op. Returns a JobResponse.
func (s *Client) UpdateAccount(ctx context.Context, opt UpdateAccountOptions) (response JobResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.UpdateAccount: ServerName is required (or set via NewForServer)")
	}
	if opt.Email == "" {
		return response, fmt.Errorf("mail.UpdateAccount: Email is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/update_account.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
