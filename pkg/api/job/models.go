package job

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

// JobsService is a Service to work with API Jobs.
type JobsService struct {
	client *api.Client
}

// SetClient set SiteHost API client.
func (s *JobsService) SetClient(c *api.Client) *JobsService {
	s.client = c

	return s
}
