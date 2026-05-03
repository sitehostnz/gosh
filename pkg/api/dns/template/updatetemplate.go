package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateTemplate renames an existing template via
// /dns/domain_templates/update_template.json. Only the name is
// editable here; SOA-default changes need a recreate or
// UpdateTemplateDNS to rebuild zones.
func (s *Client) UpdateTemplate(ctx context.Context, request UpdateTemplateRequest) (response models.APIResponse, err error) {
	if request.TemplateID == 0 {
		return response, fmt.Errorf("template.UpdateTemplate: TemplateID is required")
	}
	if request.NewName == "" {
		return response, fmt.Errorf("template.UpdateTemplate: NewName is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", strconv.Itoa(request.TemplateID))
	values.Add("new_name", request.NewName)

	req, err := s.client.NewRequest("POST", "dns/domain_templates/update_template.json",
		net.Encode(values, []string{"client_id", "template_id", "new_name"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
