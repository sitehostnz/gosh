package server

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// SetUpdateWindow configures the maintenance/patching window for
// the named CCS via "cloud/server/set_update_window.json". Returns
// a scheduler job; SDK consumers should poll until state=Completed.
//
// All five fields (ServerName, Enabled, DayOfWeek, HourOfDay,
// MinuteOfHour) are required — the API rejects with "Could not
// find a default value for the parameter '<field>'" if any are
// omitted.
func (s *Client) SetUpdateWindow(ctx context.Context, request SetUpdateWindowRequest) (response JobResponse, err error) {
	u := "cloud/server/set_update_window.json"
	keys := []string{
		"client_id",
		"server_name",
		"enabled",
		"day_of_week",
		"hour_of_day",
		"minute_of_hour",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("enabled", strconv.Itoa(request.Enabled))
	values.Add("day_of_week", strconv.Itoa(request.DayOfWeek))
	values.Add("hour_of_day", strconv.Itoa(request.HourOfDay))
	values.Add("minute_of_hour", strconv.Itoa(request.MinuteOfHour))

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
