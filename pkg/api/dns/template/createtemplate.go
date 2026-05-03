package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// CreateTemplate registers a new DNS domain template via
// /dns/domain_templates/create_template.json. Name is required;
// the rest of the fields configure SOA defaults applied to every
// domain linked to the template.
//
// Note the capitalised params[Nameserver] / params[Email] etc. —
// matches the API's published parameter casing.
//
// Empirical floor: Min must be ≥ 3600 (1 hour). The API rejects
// lower values with "Please specify a valid minimum value above
// 3600 (1 hour)." even though the public docs only list 3600 as
// an example value, not a documented floor.
func (s *Client) CreateTemplate(ctx context.Context, request CreateTemplateRequest) (response CreateTemplateResponse, err error) {
	if request.Name == "" {
		return response, fmt.Errorf("template.CreateTemplate: Name is required")
	}
	keys := []string{"client_id", "name"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)
	addOpt := func(key, val string) {
		if val == "" {
			return
		}
		values.Add(key, val)
		keys = append(keys, key)
	}
	addInt := func(key string, val int) {
		if val == 0 {
			return
		}
		values.Add(key, strconv.Itoa(val))
		keys = append(keys, key)
	}
	addOpt("params[Nameserver]", request.Nameserver)
	addOpt("params[Email]", request.Email)
	addInt("params[Refresh]", request.Refresh)
	addInt("params[Retry]", request.Retry)
	addInt("params[Expire]", request.Expire)
	addInt("params[Min]", request.Min)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/create_template.json",
		net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
