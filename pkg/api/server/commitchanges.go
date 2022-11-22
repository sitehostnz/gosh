package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// CommitChanges commit changes for upgrade a Server with the provided name.
func (s *ServersService) CommitChanges(ctx context.Context, serverName string) (*models.ServerCommitResponse, error) {
	u := "server/commit_disk_changes.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", serverName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(models.ServerCommitResponse)
	if err := s.client.Do(ctx, req, response); err != nil {
		return nil, err
	}

	return response, nil
}
