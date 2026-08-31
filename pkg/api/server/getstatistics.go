package server

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetStatistics returns metric values for one metric on one server.
//
// # Type is required
//
// Without it the API answers "The type is missing." Take the value
// from [Client.ListStatisticTypes], which enumerates both the metric
// names and the items each can be broken down by.
//
// # The item travels as options[item]
//
// Metrics that report per partition or per interface need
// [GetStatisticsOptions.Item]. It is nested under an options parent on
// the wire, so a partition or iface parameter of its own is refused
// with "One of the specified parameters is invalid, Please check your
// syntax and try again" — which does not hint that the name is the
// problem. Omitting it where it is needed gives the clearer "Partition
// Not Set".
//
// Verified live against a Xen server, August 2026.
func (s *Client) GetStatistics(ctx context.Context, opt GetStatisticsOptions) (response GetStatisticsResponse, err error) {
	if opt.ServerName == "" {
		return response, fmt.Errorf("server.GetStatistics: ServerName is required")
	}
	if opt.Type == "" {
		return response, fmt.Errorf("server.GetStatistics: Type is required; take one from ListStatisticTypes")
	}

	u := "server/get_statistics.json"
	keys := []string{"apikey", "client_id", "server_name", "type"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("server_name", opt.ServerName)
	v.Add("type", opt.Type)
	if opt.Item != "" {
		v.Add("options[item]", opt.Item)
		keys = append(keys, "options[item]")
	}
	if opt.Start != "" {
		v.Add("options[start]", opt.Start)
		keys = append(keys, "options[start]")
	}
	if opt.End != "" {
		v.Add("options[end]", opt.End)
		keys = append(keys, "options[end]")
	}
	if opt.Compacted {
		v.Add("options[compacted]", "1")
		keys = append(keys, "options[compacted]")
	}
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
