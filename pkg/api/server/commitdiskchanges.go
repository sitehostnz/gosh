package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// CommitDiskChanges function commits changes to upgrade a server.
func (s *Client) CommitDiskChanges(ctx context.Context, request CommitDiskChangesRequest) (response CommitDiskChangesResponse, err error) {
	u := "server/commit_disk_changes.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.ServerName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
