package template

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Get fetches one template's metadata + SOA defaults via
// /dns/domain_templates/get_template.json.
//
// The endpoint returns a single-element array; consumers typically
// just want response.Return[0].
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	if request.TemplateID == 0 {
		return response, fmt.Errorf("template.Get: TemplateID is required")
	}
	keys := []string{"apikey", "client_id", "template_id"}

	req, err := s.client.NewRequest("GET", "dns/domain_templates/get_template.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("template_id", strconv.Itoa(request.TemplateID))
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// ListRecords lists every record under a template via
// /dns/domain_templates/list_records.json.
func (s *Client) ListRecords(ctx context.Context, request ListRecordsRequest) (response ListRecordsResponse, err error) {
	if request.TemplateID == 0 {
		return response, fmt.Errorf("template.ListRecords: TemplateID is required")
	}
	keys := []string{"apikey", "client_id", "template_id"}

	req, err := s.client.NewRequest("GET", "dns/domain_templates/list_records.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("template_id", strconv.Itoa(request.TemplateID))
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
