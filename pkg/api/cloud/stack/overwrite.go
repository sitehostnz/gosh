package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Overwrite replaces a destination stack's contents with a source
// stack's via "cloud/stack/overwrite.json". SourceServer, Name (the
// source stack), DestinationServer, and DestinationStack are all
// required.
//
// Destructive on the destination: the destination stack's
// docker_compose is replaced with the source's. Returns a scheduler
// job id.
func (s *Client) Overwrite(ctx context.Context, request OverwriteRequest) (response JobResponse, err error) {
	uri := "cloud/stack/overwrite.json"
	keys := []string{
		"client_id",
		"source_server",
		"name",
		"destination_server",
		"destination_stack",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("source_server", request.SourceServer)
	values.Add("name", request.Name)
	values.Add("destination_server", request.DestinationServer)
	values.Add("destination_stack", request.DestinationStack)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
