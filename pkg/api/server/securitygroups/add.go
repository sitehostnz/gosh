package securitygroups

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add creates a new security group.
func (s *Client) Add(ctx context.Context, request AddRequest) (response models.APIResponse, err error) {
	uri := "server/security_groups/add.json"

	keys := []string{
		"client_id",
		"label",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", request.Label)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
