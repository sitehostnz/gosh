package stack

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// stopStartRestart is the common function for stop, start and restart actions.
func (s *Client) stopStartRestart(ctx context.Context, request StopStartRestartRequest, uri string) (response StartStopRestartResponse, err error) {
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"name",
		"containers[]",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("name", request.Name)
	for _, container := range request.Containers {
		v.Add("containers[]", container)
	}

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
