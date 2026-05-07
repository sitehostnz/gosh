package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// SearchAliases returns mail aliases on the specified mail
// server matching the given filters via
// "mail/search_aliases.json". ServerName (or NewForServer
// default) is required, plus at least one of Source or
// Destination — the API rejects calls with no query[*] filter.
func (s *Client) SearchAliases(ctx context.Context, opt SearchAliasesOptions) (response SearchAliasesResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.SearchAliases: ServerName is required (or set via NewForServer)")
	}
	if opt.Source == "" && opt.Destination == "" {
		return response, fmt.Errorf("mail.SearchAliases: at least one filter (Source / Destination) is required")
	}
	opt.ServerName = server

	u := "mail/search_aliases.json"

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
