package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetStatistics returns the metric values for the named server.
// Use ListStatisticTypes to enumerate the metric types available.
//
// Note: the parameter is "server_name" (not "name"), matching
// ListStatisticTypes — this is an API-side inconsistency relative
// to GetState/ListUpgrades/etc. on the same package.
func (s *Client) GetStatistics(ctx context.Context, opt GetStatisticsOptions) (response GetStatisticsResponse, err error) {
	u := "server/get_statistics.json"
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
