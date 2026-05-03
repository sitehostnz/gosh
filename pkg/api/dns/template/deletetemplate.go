package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// DeleteTemplate removes a template via
// /dns/domain_templates/delete_template.json. The API rejects the
// call if any domain is still linked to the template — unlink them
// first via UpdateDomain.
func (s *Client) DeleteTemplate(ctx context.Context, request DeleteTemplateRequest) (response models.APIResponse, err error) {
	if request.TemplateID == 0 {
		return response, fmt.Errorf("template.DeleteTemplate: TemplateID is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", strconv.Itoa(request.TemplateID))

	req, err := s.client.NewRequest("POST", "dns/domain_templates/delete_template.json",
		net.Encode(values, []string{"client_id", "template_id"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
