package server

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// DeleteResponse represents a result of a delete Server call.
	DeleteResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// CommitDiskChangesResponse represents a result of a commit changes Server call.
	CommitDiskChangesResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// CreateResponse represents a result of the create a Server call.
	CreateResponse struct {
		Return struct {
			JobID    string   `json:"job_id"`
			Name     string   `json:"name"`
			Password string   `json:"password"`
			Ips      []string `json:"ips"`
			ServerID string   `json:"server_id"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse represents a result of a get Server call.
	GetResponse struct {
		Server models.Server `json:"return"`
		models.APIResponse
	}

	// ListResponse lists all servers.
	ListResponse struct {
		Return struct {
			models.Pagination
			Servers []models.Server `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse represents a result of a update Server call.
	UpdateResponse struct {
		models.APIResponse
	}

	// UpgradeResponse represents a result of a upgrade Server call.
	UpgradeResponse struct {
		models.APIResponse
	}
)
