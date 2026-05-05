package template

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// CloneTemplate duplicates an existing template under a new name
// via /dns/domain_templates/clone_template.json. The clone
// inherits records and SOA defaults from the source template.
func (s *Client) CloneTemplate(ctx context.Context, request CloneTemplateRequest) (response CloneTemplateResponse, err error) {
	if request.TemplateID == "" {
		return response, fmt.Errorf("template.CloneTemplate: TemplateID is required")
	}
	if request.NewTemplateName == "" {
		return response, fmt.Errorf("template.CloneTemplate: NewTemplateName is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", request.TemplateID)
	values.Add("new_template_name", request.NewTemplateName)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/clone_template.json",
		net.Encode(values, []string{"client_id", "template_id", "new_template_name"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
