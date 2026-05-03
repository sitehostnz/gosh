package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateRecord replaces a template record in-place via
// /dns/domain_templates/update_record.json. All fields required.
func (s *Client) UpdateRecord(ctx context.Context, request UpdateRecordRequest) (response models.APIResponse, err error) {
	if request.TemplateID == 0 || request.RecordID == 0 {
		return response, fmt.Errorf("template.UpdateRecord: TemplateID and RecordID are required")
	}
	if request.Type == "" || request.Name == "" || request.Content == "" {
		return response, fmt.Errorf("template.UpdateRecord: Type, Name, Content are required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", strconv.Itoa(request.TemplateID))
	values.Add("record_id", strconv.Itoa(request.RecordID))
	values.Add("type", request.Type)
	values.Add("name", request.Name)
	values.Add("content", request.Content)
	values.Add("prio", strconv.Itoa(request.Priority))

	req, err := s.client.NewRequest("POST", "dns/domain_templates/update_record.json",
		net.Encode(values, []string{
			"client_id", "template_id", "record_id", "type", "name", "content", "prio",
		}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
