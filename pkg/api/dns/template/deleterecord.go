package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// DeleteRecord removes one record from a template via
// /dns/domain_templates/delete_record.json.
func (s *Client) DeleteRecord(ctx context.Context, request DeleteRecordRequest) (response models.APIResponse, err error) {
	if request.TemplateID == "" || request.RecordID == 0 {
		return response, fmt.Errorf("template.DeleteRecord: TemplateID and RecordID are required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("template_id", request.TemplateID)
	values.Add("record_id", strconv.Itoa(request.RecordID))

	req, err := s.client.NewRequest("POST", "dns/domain_templates/delete_record.json",
		net.Encode(values, []string{"client_id", "template_id", "record_id"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
