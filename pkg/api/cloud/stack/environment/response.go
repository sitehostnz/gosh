package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetResponse struct {
		EnvironmentVariables *[]models.EnvironmentVariable `json:"return"`
		models.APIResponse
	}
)
