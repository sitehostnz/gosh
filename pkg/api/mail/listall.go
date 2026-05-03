package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListAll returns every email-address-shaped record on the named
// mail server + domain via "mail/list_all.json" — mailboxes,
// nicknames (aliases), and forwarders, in a single flat list.
//
// Distinct from ListAccounts (mailboxes only), ListAliases, and
// ListForwards: this is the union view, with each entry's Type
// distinguishing the kind:
//
//	type=0  mailbox      Username + EmailAddr + Label
//	type=1  alias        EmailAddr + Destination
//	type=2  forward      EmailAddr + Destination
//
// Useful for one-shot account-wide audits where you just want
// "everything pointing at this domain". For typed iteration use the
// per-kind list endpoints.
func (s *Client) ListAll(ctx context.Context, opt ListAllOptions) (response ListAllResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.ListAll: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.ListAll: Domain is required")
	}
	opt.ServerName = server

	u := "mail/list_all.json"
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
