package domain

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

func (s *Client) Create(ctx context.Context, domain *models.Domain) (*models.Domain, error) {
	u := "dns/create_domain.json"

	keys := []string{
		"client_id",
		"domain",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", domain.Name)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(models.APIResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	domain = &models.Domain{
		Name:       domain.Name,
		ClientID:   s.client.ClientID,
		TemplateID: "0",
	}

	return domain, nil
}
