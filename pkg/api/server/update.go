package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Update updates a Server with the provided name.
func (s *ServersService) Update(ctx context.Context, opts *models.ServerUpdateRequest) error {
	u := "server/update.json"

	keys := []string{
		"client_id",
		"name",
		"updates[label]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("updates[label]", opts.Label)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return err
	}

	return s.client.Do(ctx, req, nil)
}
