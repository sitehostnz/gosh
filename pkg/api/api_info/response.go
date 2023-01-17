package api_info

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ApiInfoResponse struct {
		ApiInfo *models.APIInfo `json:"return"`
		models.APIResponse
	}
)
