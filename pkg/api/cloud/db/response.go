package db

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetResponse struct {
		Database models.Database `json:"return"`
		models.APIResponse
	}
	// ListResponse is the returned response from the List endpoint
	ListResponse struct {
		Return struct {
			Databases []models.Database `json:"data"`
			models.Pagination
		} `json:"return"`
		models.APIResponse
	}
)
