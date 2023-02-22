package stack

import (
	"context"
)

// Start starts a cloud stack on a given server.
func (s *Client) Start(ctx context.Context, request StopStartRestartRequest) (response StartStopRestartResponse, err error) {
	uri := "cloud/stack/start.json"
	return s.stopStartRestart(ctx, request, uri)
}
