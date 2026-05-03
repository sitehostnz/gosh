package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// SearchAccounts returns mail accounts on the specified mail
// server matching the given filters via
// "mail/search_accounts.json". ServerName (or NewForServer
// default) is required, plus at least one of EmailAddr,
// Username, Active, or Quota — the API rejects calls with no
// query[*] filter.
func (s *Client) SearchAccounts(ctx context.Context, opt SearchAccountsOptions) (response SearchAccountsResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.SearchAccounts: ServerName is required (or set via NewForServer)")
	}
	if opt.EmailAddr == "" && opt.Username == "" && opt.Active == "" && opt.Quota == "" {
		return response, fmt.Errorf("mail.SearchAccounts: at least one filter (EmailAddr / Username / Active / Quota) is required")
	}
	opt.ServerName = server

	u := "mail/search_accounts.json"

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
