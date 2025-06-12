package db

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// AddResponse is the return for adding a cloud database.
	AddResponse struct {
		Return struct {
			Job models.JobResponse
		} `json:"return"`
		models.APIResponse
	}
	// UpdateResponse is the response from updating a database.
	UpdateResponse struct {
		models.APIResponse
	}
	// DeleteResponse is the return from deleting.
	DeleteResponse struct {
		Return struct {
			Job models.JobResponse
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
