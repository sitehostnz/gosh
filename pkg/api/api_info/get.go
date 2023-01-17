package api_info

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) Get(ctx context.Context) (*models.APIInfo, error) {

	req, err := s.client.NewRequest("GET", "api/get_info.json", "")

	if err != nil {
		return nil, err
	}

	response := new(ApiInfoResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.ApiInfo, nil
}
