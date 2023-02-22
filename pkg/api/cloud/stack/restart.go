package stack

import (
	"context"
)

// Restart restarts a stack on the given server.
func (s *Client) Restart(ctx context.Context, request StopStartRestartRequest) (response StartStopRestartResponse, err error) {
	uri := "cloud/stack/restart.json"
	return s.StopStartRestart(ctx, request, uri)
}
