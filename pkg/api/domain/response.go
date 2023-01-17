package domain

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListResponse struct {
		Return struct {
			Domains *[]models.Domain `json:"data"`
		}
		models.APIResponse
	}
)
