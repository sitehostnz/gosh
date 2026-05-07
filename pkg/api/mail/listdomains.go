package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListDomains returns the mail domains configured on the
// specified mail server via "mail/list_domains.json". ServerName
// is required, either per-call in opt or via NewForServer.
//
// Each entry includes the domain, optional parent domain (for
// alias domains), catch-all address, state, and per-domain
// counts of accounts, nicknames (aliases), forwarders, and total
// addresses in use.
func (s *Client) ListDomains(ctx context.Context, opt ListDomainsOptions) (response ListDomainsResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.ListDomains: ServerName is required (or set via NewForServer)")
	}
	opt.ServerName = server

	u := "mail/list_domains.json"

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
