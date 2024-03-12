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
		Return struct {
			ID          string `json:"id"`
			ClientID    string `json:"client_id"`
			Label       string `json:"label"`
			Content     string `json:"content"`
			DateAdded   string `json:"date_added"`
			DateUpdated string `json:"date_updated"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse represents a result of the delete an SSH Key call.
	DeleteResponse struct {
		models.APIResponse
	}
)
