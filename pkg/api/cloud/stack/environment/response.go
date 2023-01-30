package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetResponse struct {
		EnvironmentVariables *[]models.EnvironmentVariable `json:"return"`
		models.APIResponse
	}
	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
