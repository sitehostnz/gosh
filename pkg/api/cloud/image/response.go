package image

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the return from the /cloud/image/list_all.json endpoint.
	ListResponse struct {
		Return struct {
			models.Pagination
			Images []models.CloudImage `json:"data"`
		}
		models.APIResponse
	}

	// GetResponse represents the return from the /cloud/image/get.json endpoint.
	GetResponse struct {
		Image models.CloudImage `json:"return"`
		models.APIResponse
	}

	// JobResponse is the shared response for image write operations
	// that queue a scheduler job (Create, Delete).
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// Changelog is the body of GetChangelogResponse.
	Changelog struct {
		Code      string `json:"code"`
		Label     string `json:"label"`
		Changelog string `json:"changelog"`
	}

	// GetChangelogResponse represents the return from
	// /cloud/image/get_changelog.json.
	GetChangelogResponse struct {
		Return Changelog `json:"return"`
		models.APIResponse
	}
)
