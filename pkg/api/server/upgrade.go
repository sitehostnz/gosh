package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Upgrade upgrades a Server.
func (s *ServersService) Upgrade(ctx context.Context, opts *models.ServerUpgradeRequest) error {
	u := "server/upgrade_plan.json"

	keys := []string{
		"client_id",
		"name",
		"plan",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("plan", opts.Plan)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return err
	}

	return s.client.Do(ctx, req, nil)
}
