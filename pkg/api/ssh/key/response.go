package key

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the listing of SSHKeys.
	ListResponse struct {
		Return struct {
			SSHKeys []models.SSHKey `json:"data"`
		} `json:"return"`
		models.APIResponse
	}
)
