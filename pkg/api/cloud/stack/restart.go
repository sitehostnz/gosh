package stack

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Restart restarts a stack on the given server
func (s *Client) Restart(ctx context.Context, request StopStartRequest) (*StartStopResponse, error) {

	u := "cloud/stack/restart.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"name",
	}
	// leave out the containers for now...

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("name", request.Name)

	req.URL.RawQuery = utils.Encode(v, keys)

	response := new(StartStopResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
