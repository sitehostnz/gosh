package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Copy duplicates an existing cloud stack onto a destination server
// via "cloud/stack/copy.json". SourceServer, Name, DestinationServer,
// and Label are required. The new stack inherits the source's
// docker_compose; Label sets the new stack's label.
//
// SourceServer and DestinationServer may be the same when copying
// within a single server. Returns a scheduler job id.
func (s *Client) Copy(ctx context.Context, request CopyRequest) (response JobResponse, err error) {
	uri := "cloud/stack/copy.json"
	keys := []string{
		"client_id",
		"source_server",
		"name",
		"destination_server",
		"label",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("source_server", request.SourceServer)
	values.Add("name", request.Name)
	values.Add("destination_server", request.DestinationServer)
	values.Add("label", request.Label)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
