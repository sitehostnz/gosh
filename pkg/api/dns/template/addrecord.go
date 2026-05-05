package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// AddRecord adds a record to a template via
// /dns/domain_templates/add_record.json.
//
// Type must be one of: A, AAAA, NS, MX, PTR, SRV, TXT, CNAME.
// Priority is only meaningful for MX/SRV; pass 0 otherwise.
func (s *Client) AddRecord(ctx context.Context, request AddRecordRequest) (response models.APIResponse, err error) {
	if request.TemplateID == "" {
		return response, fmt.Errorf("template.AddRecord: TemplateID is required")
	}
	if request.Type == "" || request.Name == "" || request.Content == "" {
		return response, fmt.Errorf("template.AddRecord: Type, Name, Content are required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", request.TemplateID)
	values.Add("type", request.Type)
	values.Add("name", request.Name)
	values.Add("content", request.Content)
	values.Add("prio", strconv.Itoa(request.Priority))

	req, err := s.client.NewRequest("POST", "dns/domain_templates/add_record.json",
		net.Encode(values, []string{
			"client_id", "template_id", "type", "name", "content", "prio",
		}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
