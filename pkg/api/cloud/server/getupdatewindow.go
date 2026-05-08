package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetUpdateWindow reads the maintenance/patching window configuration
// for the named CCS via "cloud/server/get_update_window.json".
//
// Note: this endpoint returns "This server is not managed by SiteHost"
// for managed=0 servers — only managed CCSes have a configured window.
func (s *Client) GetUpdateWindow(ctx context.Context, request GetUpdateWindowRequest) (response GetUpdateWindowResponse, err error) {
	u := "cloud/server/get_update_window.json"
	keys := []string{"apikey", "client_id", "server_name"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("server_name", request.ServerName)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
