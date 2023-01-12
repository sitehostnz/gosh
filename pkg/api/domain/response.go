package domain

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListResponse struct {
		//Server models.Server `json:"return"`
		Return struct {
			Domains *[]models.Domain `json:"data"`
		}
		models.APIResponse
	}
)
