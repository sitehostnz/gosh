package mail

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// AddForward adds a forwarder mapping (Source → Destination)
// via "mail/add_forward.json". ServerName, Source, and
// Destination are required. Returns a JobResponse.
func (s *Client) AddForward(ctx context.Context, opt AddForwardOptions) (response JobResponse, err error) {
	server := s.resolveServerName(opt.ServerName)
	if server == "" {
		return response, fmt.Errorf("mail.AddForward: ServerName is required (or set via NewForServer)")
	}
	if opt.Source == "" {
		return response, fmt.Errorf("mail.AddForward: Source is required")
	}
	if opt.Destination == "" {
		return response, fmt.Errorf("mail.AddForward: Destination is required")
	}
	opt.ServerName = server

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "mail/add_forward.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
