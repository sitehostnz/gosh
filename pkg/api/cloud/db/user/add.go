package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add creates a new cloud database.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/db/user/add.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"username",
		"password",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("password", request.Password)

	// suspect these are required at the same time, so lets not deal with this here?
	// feels cleaner to handle a grant independantly
	//if request.Database != "" {
	//	values.Add("database", request.Database)
	//}
	//if request.Grants != nil {
	//}

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
