package api_info

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ApiInfoResponse struct {
		ApiInfo *models.ApiInfo `json:"return"`
		models.APIResponse
	}
)
