package grant

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// AddResponse represents a response from adding a database grant with the `/cloud/db/grant/add.json` endpoint.
	AddResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
	// UpdateResponse represents a response from updating a database grant with the `/cloud/db/grant/update.json` endpoint.
	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse represents a response from deleting a database grant with the `/cloud/db/grant/delete.json` endpoint.
	DeleteResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
