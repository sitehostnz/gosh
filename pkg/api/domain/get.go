package domain

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) Get(ctx context.Context, request GetRequest) (*models.Domain, error) {
	u := fmt.Sprintf("dns/list_domains.json?filters[domain]=%v", request.DomainName)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	if len(*response.Return.Domains) == 0 {
		return nil, nil
	}
	// we're assuming that when we ask for a domain, the list here should be unique and we get one or none
	return &(*response.Return.Domains)[0], nil
}
