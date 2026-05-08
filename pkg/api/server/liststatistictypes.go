package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListStatisticTypes returns the metric types available for the
// named server (passed to GetStatistics).
//
// Note: this endpoint uses "server_name" as the parameter, distinct
// from sibling endpoints like GetState that use plain "name". The
// distinction is at the API level — the wrapper just transmits.
func (s *Client) ListStatisticTypes(ctx context.Context, opt ListStatisticTypesOptions) (response ListStatisticTypesResponse, err error) {
	u := "server/list_statistic_types.json"
	keys := []string{"apikey", "client_id", "server_name"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("server_name", opt.ServerName)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
