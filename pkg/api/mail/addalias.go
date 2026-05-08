package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// AddAlias adds an alias mapping (Source → Destination) via
// "mail/add_alias.json". ServerName, Source, and Destination
// are required. Returns a JobResponse.
func (s *Client) AddAlias(ctx context.Context, opt AddAliasOptions) (response JobResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.AddAlias: ServerName is required (or set via NewForServer)")
	}
	if opt.Source == "" {
		return response, fmt.Errorf("mail.AddAlias: Source is required")
	}
	if opt.Destination == "" {
		return response, fmt.Errorf("mail.AddAlias: Destination is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/add_alias.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
