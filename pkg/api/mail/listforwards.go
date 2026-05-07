package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListForwards returns the mail forwarders for a domain on the
// specified mail server via "mail/list_forwards.json". Both
// ServerName (or NewForServer default) and Domain are required;
// Source is an optional filter narrowing to a specific source
// address.
func (s *Client) ListForwards(ctx context.Context, opt ListForwardsOptions) (response ListForwardsResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.ListForwards: ServerName is required (or set via NewForServer)")
	}
	if opt.Domain == "" {
		return response, fmt.Errorf("mail.ListForwards: Domain is required")
	}
	opt.ServerName = server

	u := "mail/list_forwards.json"

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
