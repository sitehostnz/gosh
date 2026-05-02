package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Backup creates a backup of a cloud stack via
// "cloud/stack/backup.json". ServerName and Name are required.
// Returns a scheduler job id; the backup is taken asynchronously.
func (s *Client) Backup(ctx context.Context, request BackupRequest) (response JobResponse, err error) {
	uri := "cloud/stack/backup.json"
	keys := []string{"client_id", "server", "name"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
