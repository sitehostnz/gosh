package image

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the return from the /cloud/images/list_all.json endpoint.
	ListResponse struct {
		Return struct {
			models.Pagination
			Images []models.Image `json:"data"`
		}
		models.APIResponse
	}

	// GetResponse represents the return from the /cloud/images/get.json endpoint.
	GetResponse struct {
		Image models.Image `json:"return"`
		models.APIResponse
	}
)
