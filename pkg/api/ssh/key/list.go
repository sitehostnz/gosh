package sshkey

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
)

// List returns a list of all ssh keys.
func (s *Client) List(ctx context.Context) (*[]models.SSHKey, error) {
	req, err := s.client.NewRequest("GET", "ssh/key/list_all.json", "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// if the request works, don't care about pagination right now.
	return response.Return.SSHKeys, err
}
