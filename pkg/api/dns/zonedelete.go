package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// DeleteZone deletes an existing DNS zone with the specified domain name.
// It takes a context.Context and a DeleteZoneRequest struct as input parameters.
// It returns a models.APIResponse pointer and an error.
func (s *Client) DeleteZone(ctx context.Context, request DeleteZoneRequest) (response models.APIResponse, err error) {
	u := "dns/delete_domain.json"

	keys := []string{
		"client_id",
		"domain",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", request.DomainName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
