package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// DeleteAccount removes a mail account via
// "mail/delete_account.json". ServerName and Email are required.
// Returns a JobResponse.
func (s *Client) DeleteAccount(ctx context.Context, opt DeleteAccountOptions) (response JobResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.DeleteAccount: ServerName is required (or set via NewForServer)")
	}
	if opt.Email == "" {
		return response, fmt.Errorf("mail.DeleteAccount: Email is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/delete_account.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
