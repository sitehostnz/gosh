package stack

import (
	"context"
)

// Stop is for stopping a cloud stack on a given server.
func (s *Client) Stop(ctx context.Context, request StopStartRestartRequest) (response StartStopRestartResponse, err error) {
	uri := "cloud/stack/stop.json"
	return s.stopStartRestart(ctx, request, uri)
}
