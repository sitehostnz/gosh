package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List returns cloud database users, via
// "cloud/db/user/list_all.json".
//
// Every filter is optional: with no options this returns every database
// user on the account. As with [db.Client.List], a server name that does
// not resolve is rejected rather than silently ignored, so an empty
// page means the filter matched nothing real.
//
// The password field in each result is always empty. Passwords are
// write-only; [Client.Update] sets one, nothing reads one back.
func (s *Client) List(ctx context.Context, opt ListOptions) (response ListResponse, err error) {
	uri := "cloud/db/user/list_all.json"

	path, err := net.AddOptions(uri, opt)
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
