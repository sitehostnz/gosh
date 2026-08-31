package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List returns cloud SSH users, via "cloud/ssh/user/list_all.json".
//
// Every filter is optional: with no options this returns every SSH user
// on the account. A server name that does not resolve is rejected
// rather than ignored.
//
// Each result carries the user's home directory and the SSH keys
// attached to it, so this is the call that answers "which keys can
// reach this container".
func (s *Client) List(ctx context.Context, options ListOptions) (response ListResponse, err error) {
	uri := "cloud/ssh/user/list_all.json"

	path, err := net.AddOptions(uri, options)
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
