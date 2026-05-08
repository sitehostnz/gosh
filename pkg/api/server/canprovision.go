package server

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// CanProvision checks resource availability for provisioning a
// server product via "server/can_provision.json". Product,
// Location, and Distro are required; Arch is optional.
// Synchronous; returns models.APIResponse — Status indicates
// whether the resources are available.
func (s *Client) CanProvision(ctx context.Context, opt CanProvisionOptions) (response models.APIResponse, err error) {
	if opt.Product == "" {
		return response, fmt.Errorf("server.CanProvision: Product is required")
	}
	if opt.Location == "" {
		return response, fmt.Errorf("server.CanProvision: Location is required")
	}
	if opt.Distro == "" {
		return response, fmt.Errorf("server.CanProvision: Distro is required")
	}

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "server/can_provision.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
