package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get fetches a cloud db.
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

	v := req.URL.Query()
	v.Add("api_key", s.client.APIKey)
	v.Add("client_id", s.client.ClientID)
	v.Add("server_name", request.ServerName)
	v.Add("mysql_host", request.MySQLHost)
	v.Add("username", request.Username)

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
