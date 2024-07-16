package image

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the return from the /cloud/stack/images/list_all.json endpoint.
	ListResponse struct {
		Return []models.StackImage `json:"return"`
		models.APIResponse
	}
)
