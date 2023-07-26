package user

import (
	"github.com/sitehostnz/gosh/pkg/models"
)

type (
	// ListResponse is the returned response from the List endpoint.
	ListResponse struct {
		Return struct {
			Users []models.DatabaseUser `json:"data"`
			models.Pagination
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse gets the specified database user.
	GetResponse struct {
		User models.DatabaseUser `json:"return"`
		models.APIResponse
	}

	// UpdateResponse updates the specified database user.
	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// AddResponse adds a new database user.
	AddResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse deletes the specified database user.
	DeleteResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
