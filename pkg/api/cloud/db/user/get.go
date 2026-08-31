package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Get fetches a cloud db user.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/db/user/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server_name",
		"mysql_host",
		"username",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	// apikey and client_id are already on the query from NewRequest.
	// Re-adding them sent client_id twice, and "api_key" is not a
	// parameter this API has — it was absent from the keys list, so
	// net.Encode dropped it and nothing ever complained.
	v := req.URL.Query()
	v.Add("server_name", request.ServerName)
	v.Add("mysql_host", request.MySQLHost)
	v.Add("username", request.Username)

	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
