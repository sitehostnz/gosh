package server

import "github.com/sitehostnz/gosh/pkg/models"

// DeleteResponse represents a result of a delete Server call.
type DeleteResponse struct {
	Return struct {
		JobID string `json:"job_id"`
	} `json:"return"`
	models.APIResponse
}

// CommitDiskChangesResponse represents a result of a commit changes Server call.
type CommitDiskChangesResponse struct {
	Return struct {
		JobID string `json:"job_id"`
	} `json:"return"`
	models.APIResponse
}

// CreateResponse represents a request to create a Server.
type CreateResponse struct {
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
type GetResponse struct {
	Server models.Server `json:"return"`
	models.APIResponse
}

// UpdateResponse represents a result of a update Server call.
type UpdateResponse struct {
	models.APIResponse
}

// UpgradeResponse represents a result of a upgrade Server call.
type UpgradeResponse struct {
	models.APIResponse
}
