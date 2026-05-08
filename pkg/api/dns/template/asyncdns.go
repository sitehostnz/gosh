package template

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateDomainDNS forces a rebuild of one domain's zone from its
// currently-linked template via
// /dns/domain_templates/update_domain_dns.json. Returns a scheduler
// job; poll /job/get to track completion.
//
// Useful after editing a template's records, when you want the
// change applied to a specific domain immediately rather than
// waiting for the bulk UpdateTemplateDNS rebuild.
func (s *Client) UpdateDomainDNS(ctx context.Context, request UpdateDomainDNSRequest) (response JobResponse, err error) {
	if request.Domain == "" {
		return response, fmt.Errorf("template.UpdateDomainDNS: Domain is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", request.Domain)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/update_domain_dns.json",
		net.Encode(values, []string{"client_id", "domain"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UpdateTemplateDNS forces a rebuild of every domain linked to the
// given template via
// /dns/domain_templates/update_template_dns.json. Returns a
// scheduler job that may take a while if the template has many
// linked domains.
//
// The bulk equivalent of UpdateDomainDNS — use this after editing
// records on a template when you want the change applied to every
// domain it serves.
func (s *Client) UpdateTemplateDNS(ctx context.Context, request UpdateTemplateDNSRequest) (response JobResponse, err error) {
	if request.TemplateID == "" {
		return response, fmt.Errorf("template.UpdateTemplateDNS: TemplateID is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", request.TemplateID)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/update_template_dns.json",
		net.Encode(values, []string{"client_id", "template_id"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
