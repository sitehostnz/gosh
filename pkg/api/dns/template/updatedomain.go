package template

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateDomain points a domain at a different template via
// /dns/domain_templates/update_domain.json. Pass TemplateID="" to
// unlink the domain entirely (it then keeps its current zone but
// no longer tracks any template).
func (s *Client) UpdateDomain(ctx context.Context, request UpdateDomainRequest) (response UpdateDomainResponse, err error) {
	if request.Domain == "" {
		return response, fmt.Errorf("template.UpdateDomain: Domain is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", request.Domain)
	values.Add("params[template_id]", request.TemplateID)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/update_domain.json",
		net.Encode(values, []string{"client_id", "domain", "params[template_id]"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
