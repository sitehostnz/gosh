package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse is the response that returns a stacks environment variables.
	GetResponse struct {
		EnvironmentVariables []models.EnvironmentVariable `json:"return"`
		models.APIResponse
	}

	// UpdateResponse is the result of updating an environment on a stack.
	UpdateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
