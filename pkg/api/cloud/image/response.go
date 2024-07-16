package image

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the return from the /cloud/stack/image/list_all.json endpoint.
	ListResponse struct {
		Return struct {
			models.Pagination
			Images []models.CloudImage `json:"data"`
		}
		models.APIResponse
	}

	// GetResponse represents the return from the /cloud/images/get.json endpoint.
	GetResponse struct {
		Image models.CloudImage `json:"return"`
		models.APIResponse
	}
)
