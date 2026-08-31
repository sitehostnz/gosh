package stack

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Get fetches a single stack, via "cloud/stack/get.json".
//
// # The parameter is "server", not "server_name"
//
// Every other stack and database endpoint identifies the server with
// server_name (or filters[server_name]). This one wants server. Sending
// server_name instead is rejected with "The server name is missing.",
// which reads like the value was omitted rather than misnamed.
//
// [GetRequest.ServerName] is therefore named for consistency with the
// rest of the package, not for the wire; the mapping happens below.
//
// Name is the stack's name, not its label — the short generated
// identifier ("cc0123456789abcd") or a chosen one ("infra"), as
// returned in the name field of [Client.List].
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/stack/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"name",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("name", request.Name)

	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
