package key

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the listing of SSHKeys.
	ListResponse struct {
		Return struct {
			models.Pagination
			SSHKeys []models.SSHKey `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// CreateResponse represents a result of the create an SSH Key call.
	CreateResponse struct {
		Return struct {
			KeyID string `json:"key_id"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse represents a SSH Keys information.
	GetResponse struct {
		Return models.SSHKey `json:"return"`
		models.APIResponse
	}

	// DeleteResponse represents a result of a delete an SSH Key call.
	DeleteResponse struct {
		models.APIResponse
	}

	// UpdateResponse represents a result of the update an SSH Key call.
	UpdateResponse struct {
		models.APIResponse
	}
)
