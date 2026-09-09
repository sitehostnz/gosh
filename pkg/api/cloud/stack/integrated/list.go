package integrated

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List returns the names of the integrated stacks on a server, via
// "cloud/stack/integrated/list_all.json".
//
// The names it returns are exactly what cloud/db endpoints want as
// their MySQLHost, so this is the call that answers "where can I put a
// database on this server". An empty list means nowhere yet — see
// [Client.Add].
//
// # Use this rather than filtering cloud/stack/list_all.json
//
// Integrated stacks are not reliably present in the general stack
// listing: on one server a database stack was absent from it while
// demonstrably running and accepting databases, and present on
// another. Prefix-matching that listing for "mysql" or "mariadb" is
// therefore a discovery method that works until it does not.
//
// # The response shape is its own
//
// The return is a bare array of strings, with no pagination and no
// data wrapper, unlike every other cloud listing.
func (s *Client) List(ctx context.Context, request ListRequest) (response ListResponse, err error) {
	if request.ServerName == "" {
		return response, fmt.Errorf("cloud/stack/integrated.List: ServerName is required")
	}

	req, err := s.client.NewRequest("GET", "cloud/stack/integrated/list_all.json", "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server_name", request.ServerName)
	req.URL.RawQuery = net.Encode(v, []string{"apikey", "client_id", "server_name"})

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
