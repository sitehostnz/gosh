package template

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// SearchTemplates fuzzy-matches templates by name via
// /dns/domain_templates/search_templates.json. TemplateName is
// required; Offset and Limit are optional pagination knobs.
//
// Results from SearchTemplates have a different shape than
// ListTemplates — no template_id field, SOA defaults inline. See
// SearchResult.
//
// Empirical: matching appears to be **exact** on template_name,
// not the substring / fuzzy match the docs' phrasing implies.
// Searching for "gosh-example-" doesn't return templates named
// "gosh-example-abc12345"; only the full-name query does. Treat
// this as a lookup by exact name until the API contract clarifies.
func (s *Client) SearchTemplates(ctx context.Context, request SearchTemplatesRequest) (response SearchTemplatesResponse, err error) {
	if request.TemplateName == "" {
		return response, fmt.Errorf("template.SearchTemplates: TemplateName is required")
	}
	keys := []string{"client_id", "query[template_name]"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("query[template_name]", request.TemplateName)
	if request.Offset > 0 {
		values.Add("limitArr[offset]", strconv.Itoa(request.Offset))
		keys = append(keys, "limitArr[offset]")
	}
	if request.Limit > 0 {
		values.Add("limitArr[limit]", strconv.Itoa(request.Limit))
		keys = append(keys, "limitArr[limit]")
	}

	req, err := s.client.NewRequest("POST", "dns/domain_templates/search_templates.json",
		net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
