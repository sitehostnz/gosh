package firewall

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// UpdateResponse represents the result of a request to update a server's firewall call.
	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
