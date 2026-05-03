package mail

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetServerInfo returns connection details (hostname, webmail
// URL, etc.) for the specified mail server via
// "mail/get_server_info.json". ServerName is required, either
// per-call in opt or via NewForServer.
func (s *Client) GetServerInfo(ctx context.Context, opt GetServerInfoOptions) (response GetServerInfoResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.GetServerInfo: ServerName is required (or set via NewForServer)")
	}
	opt.ServerName = server

	u := "mail/get_server_info.json"

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
