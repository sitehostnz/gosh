package image

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse is the response that returns a images list.
	GetResponse struct {
		Return []models.Image `json:"return"`
		models.APIResponse
	}
)
