package db

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetResponse struct {
		//models.APIResponse
	}
	ListResponse struct {
		Return struct {
			Databases []models.Database `json:"data"`
			models.Pagination
		} `json:"return"`
		models.APIResponse
	}
)
