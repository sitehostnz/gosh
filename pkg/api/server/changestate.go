package server

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// State values accepted by ChangeState.
const (
	StatePowerOn   = "power_on"
	StatePowerOff  = "power_off"
	StateRescueOn  = "rescue_on"
	StateRescueOff = "rescue_off"
	StateReboot    = "reboot"
)

// ChangeState transitions a server's power / rescue state via
// "server/change_state.json". Both Name and State are required;
// State must be one of the State* constants. Returns the
// scheduler job for the state-change task.
//
// **This is destructive.** Power-off and reboot interrupt
// running workloads on the server.
func (s *Client) ChangeState(ctx context.Context, opt ChangeStateOptions) (response ChangeStateResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("server.ChangeState: Name is required")
	}
	if opt.State == "" {
		return response, fmt.Errorf("server.ChangeState: State is required (one of power_on/power_off/rescue_on/rescue_off/reboot)")
	}

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "server/change_state.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
