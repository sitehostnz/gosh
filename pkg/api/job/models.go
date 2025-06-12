package job

import (
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/models"
)

const (
	// SchedulerType is a job that runs on a scheduler.
	SchedulerType string = "scheduler"

	// DaemonType is a job that runs on a daemon.
	DaemonType = "daemon"
)

type (
	// Client is a Service to work with API Jobs.
	Client struct {
		client *api.Client
	}

	// GetRequest represents a SiteHost request for GET job api endpoint.
	GetRequest struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	}

	// GetResponse represents a SiteHost response for GET job api endpoint.
	GetResponse struct {
		Return models.JobDetails `json:"return"`
		models.APIResponse
	}
)

// New is an initialisation function.
func New(c *api.Client) *Client {
	return &Client{
		client: c,
	}
}
