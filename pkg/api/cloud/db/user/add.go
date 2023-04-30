package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add creates a new for cloud database user.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/db/user/add.json"
	keys := []string{
		"server_name",
		"mysql_host",
		"username",
		"password",
		"database",
		"grants[]",
	}

	values := url.Values{}
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("password", request.Password)
	values.Add("database", request.Database)

	// database=testdb2, grants[0]=select, grants[1]=insert, grants[2]=update, grants[3]=delete, grants[4]=create, grants[5]=drop, grants[6]=alter, grants[7]=index, grants[8]=create view, grants[9]=show view, grants[10]=lock tables, grants[11]=create temporary tables

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
