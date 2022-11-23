package job

import (
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/models"
)

// Client is a Service to work with API Jobs.
type Client struct {
	client *api.Client
}

// New is an initialisation function.
func New(c *api.Client) *Client {
	return &Client{
		client: c,
	}
}

// GetRequest represents a SiteHost request for GET job api endpoint.
type GetRequest struct {
	JobID string `json:"job_id"`
	Type  string `json:"type"`
}

// GetResponse represents a SiteHost response for GET job api endpoint.
type GetResponse struct {
	Return models.JobDetails `json:"return"`
	models.APIResponse
}
