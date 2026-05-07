package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListAccounts returns the mail accounts for a domain on the
// specified mail server via "mail/list_accounts.json". Both
// ServerName (or NewForServer default) and Domain are required;
// EmailAddr is an optional filter restricting results to a
// specific address.
func (s *Client) ListAccounts(ctx context.Context, opt ListAccountsOptions) (response ListAccountsResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.ListAccounts: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.ListAccounts: Domain is required")
	}
	opt.ServerName = server

	u := "mail/list_accounts.json"

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
