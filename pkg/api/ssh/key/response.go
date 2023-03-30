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
	// GetResponse is the return from a GetRequest.
	GetResponse struct {
		Return models.SSHKey `json:"return"`
		models.APIResponse
	}
	// AddResponse is the return from an AddRequest.
	AddResponse struct {
		Return struct {
			KeyID string `json:"key_id"`
		} `json:"return"`
		models.APIResponse
	}
	// UpdateResponse is the return from an UpdateRequest.
	UpdateResponse struct {
		models.APIResponse
	}
	// RemoveResponse is the return from an RemoveRequest.
	RemoveResponse struct {
		models.APIResponse
	}
)
