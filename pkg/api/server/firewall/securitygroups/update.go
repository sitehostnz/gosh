package securitygroups

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update updates a security group.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := path.Join(apiPrefix, "update.json")

	keys := []string{
		"client_id",
		"name",
		"params[label]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)
	values.Add("params[label]", request.Params.Label)

	for i, rulein := range request.Params.RulesIn {
		prefix := fmt.Sprintf("params[rules_in][%d]", i)
		keys = append(keys, prefix+"[enabled]")
		keys = append(keys, prefix+"[action]")
		keys = append(keys, prefix+"[protocol]")
		keys = append(keys, prefix+"[src_ip]")
		keys = append(keys, prefix+"[dest_port]")
		values.Add(prefix+"[enabled]", strconv.Itoa(types.BoolToInt(rulein.Enabled)))
		values.Add(prefix+"[action]", rulein.Action)
		values.Add(prefix+"[protocol]", rulein.Protocol)
		values.Add(prefix+"[src_ip]", rulein.IP)
		values.Add(prefix+"[dest_port]", rulein.DestinationPort)
	}

	for i, ruleout := range request.Params.RulesOut {
		prefix := fmt.Sprintf("params[rules_out][%d]", i)
		keys = append(keys, prefix+"[enabled]")
		keys = append(keys, prefix+"[action]")
		keys = append(keys, prefix+"[protocol]")
		keys = append(keys, prefix+"[dest_ip]")
		keys = append(keys, prefix+"[dest_port]")
		values.Add(prefix+"[enabled]", strconv.Itoa(types.BoolToInt(ruleout.Enabled)))
		values.Add(prefix+"[action]", ruleout.Action)
		values.Add(prefix+"[protocol]", ruleout.Protocol)
		values.Add(prefix+"[dest_ip]", ruleout.IP)
		values.Add(prefix+"[dest_port]", ruleout.DestinationPort)
	}

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
