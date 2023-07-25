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

	GetResponse struct {
		User models.DatabaseUser `json:"return"`
		models.APIResponse
	}

	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	AddResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	DeleteResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
