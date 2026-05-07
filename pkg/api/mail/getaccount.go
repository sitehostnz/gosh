package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetAccount returns the full record for a single mail account
// (identified by email address) via "mail/get_account.json".
// Both ServerName (or NewForServer default) and Email are
// required.
func (s *Client) GetAccount(ctx context.Context, opt GetAccountOptions) (response GetAccountResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.GetAccount: ServerName is required (or set via NewForServer)")
	}
	if opt.Email == "" {
		return response, fmt.Errorf("mail.GetAccount: Email is required")
	}
	opt.ServerName = server

	u := "mail/get_account.json"

	path, err := net.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
