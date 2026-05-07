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
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse is returned by environment/delete. Delete is
	// async: the API queues a job and the response carries its
	// descriptor. Poll via job.Get to confirm the variables were
	// actually removed.
	DeleteResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
