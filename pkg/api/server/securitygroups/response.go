package securitygroups

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// UpdateResponse represents a result of an update security group call.
	UpdateResponse struct {
		Return struct {
			Type  string `json:"type"`
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
