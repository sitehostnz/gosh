package redirect

import (
	"context"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListRedirects returns every redirect rule on the account via
// /redirect/list_redirects.json, grouped by domain → source URL →
// (destination + HTTP type).
//
// Pagination filters control the result window if the account has
// a lot of rules.
func (s *Client) ListRedirects(ctx context.Context, request ListRedirectsRequest) (response ListRedirectsResponse, err error) {
	keys := []string{"apikey", "client_id"}

	req, err := s.client.NewRequest("GET", "redirect/list_redirects.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	if request.SortBy != "" {
		v.Add("filters[sort_by]", request.SortBy)
		keys = append(keys, "filters[sort_by]")
	}
	if request.SortDir != "" {
		v.Add("filters[sort_dir]", request.SortDir)
		keys = append(keys, "filters[sort_dir]")
	}
	if request.PageSize != 0 {
		v.Add("filters[page_size]", strconv.Itoa(request.PageSize))
		keys = append(keys, "filters[page_size]")
	}
	if request.PageNumber != 0 {
		v.Add("filters[page_number]", strconv.Itoa(request.PageNumber))
		keys = append(keys, "filters[page_number]")
	}
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
