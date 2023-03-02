package stack

import (
	"github.com/sitehostnz/gosh/pkg/models"
)

type (
	// GetResponse GetReponse is the response for getting a single stack.
	GetResponse struct {
		Stack models.Stack `json:"return"`
		models.APIResponse
	}

	// AddResponse is the response from calling the create stack api.
	AddResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// ListResponse is the response for listing stacks.
	ListResponse struct {
		Return struct {
			models.Pagination
			Stacks []models.Stack `json:"data"`
		}
		models.APIResponse
	}

	// StartStopRestartResponse represents the response from start/stop/restart actions.
	StartStopRestartResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// GenerateNameResponse represents the response from generate_name action.
	GenerateNameResponse struct {
		Return struct {
			Name string `json:"name"`
		} `json:"return"`
		models.APIResponse
	}
)
