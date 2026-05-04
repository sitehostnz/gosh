package client

import (
	"context"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListSubAccounts returns the paginated list of sub-accounts under
// the calling client_id, via /accounts/client/list_sub_accounts.json.
//
// Useful for reseller / parent-account discovery — combined with
// info.NewClientWithDiscovery, callers can enumerate every customer
// they have visibility into without hardcoding IDs.
//
// IncludeClosed defaults to false; closed accounts (those with a
// non-zero Closed date) are excluded unless explicitly requested.
func (s *Client) ListSubAccounts(ctx context.Context, request ListSubAccountsRequest) (response ListSubAccountsResponse, err error) {
	keys := []string{"apikey", "client_id"}

	req, err := s.client.NewRequest("GET", "accounts/client/list_sub_accounts.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	if request.Name != "" {
		v.Add("filters[name]", request.Name)
		keys = append(keys, "filters[name]")
	}
	if request.IncludeClosed {
		v.Add("filters[include_closed]", "1")
		keys = append(keys, "filters[include_closed]")
	}
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
