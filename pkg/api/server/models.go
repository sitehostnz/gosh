package server

import "github.com/sitehostnz/gosh/pkg/api"

// ServersService is a Service to work with API Server.
type ServersService struct {
	client *api.Client
}

// SetClient set SiteHost API client.
func (s *ServersService) SetClient(c *api.Client) *ServersService {
	s.client = c

	return s
}
