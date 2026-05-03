package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListAliases returns the mail aliases for a domain on the
// specified mail server via "mail/list_aliases.json". Both
// ServerName (or NewForServer default) and Domain are required;
// Source is an optional filter narrowing to a specific source
// address.
func (s *Client) ListAliases(ctx context.Context, opt ListAliasesOptions) (response ListAliasesResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.ListAliases: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.ListAliases: Domain is required")
	}
	opt.ServerName = server

	u := "mail/list_aliases.json"

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
