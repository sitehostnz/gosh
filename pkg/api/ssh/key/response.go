package key

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the listing of SSH keys.
	ListResponse struct {
		Return struct {
			models.Pagination
			SSHKeys []models.SSHKey `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// CreateResponse represents a result of creating an SSH key.
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

	// DeleteResponse represents the result of deleting an SSH key.
	DeleteResponse struct {
		models.APIResponse
	}

	// UpdateResponse represents a result of updating an SSH key.
	UpdateResponse struct {
		models.APIResponse
	}
)
