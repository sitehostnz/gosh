package letsencrypt

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List retrieves the LE certificates currently provisioned for the
// named stack on a Cloud Container Server. Returns a map keyed by
// container name within that stack; containers without an LE cert
// are absent from the map. Containers (optional) restricts the
// result to specific containers.
func (s *Client) List(ctx context.Context, request ListRequest) (response ListResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/list_all.json"
	keys := []string{"apikey", "client_id", "server", "name", "containers"}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("name", request.StackName)
	for _, container := range request.Containers {
		v.Add("containers", container)
	}
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
