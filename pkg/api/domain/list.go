package domain

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) List(ctx context.Context) (*[]models.Domain, error) {
	u := fmt.Sprintf("dns/list_domains.json")
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// if the request works, don't care about pagination right now...
	return response.Return.Domains, err
}
