package user

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// AddResponse is the return for adding a cloud database user.
	AddResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse is the response from updating a database user.
	UpdateResponse struct {
		models.APIResponse
	}

	// DeleteResponse is the return from deleting database user.
	DeleteResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse is the return value from a get request.
	GetResponse struct {
		Database models.Database `json:"return"`
		models.APIResponse
	}

	// ListResponse is the returned response from the List endpoint.
	ListResponse struct {
		Return struct {
			Databases []models.Database `json:"data"`
			models.Pagination
		} `json:"return"`
		models.APIResponse
	}
)
