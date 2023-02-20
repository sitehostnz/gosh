package stack

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse GetReponse is the response for getting a single stack
	GetResponse struct {
		Stack *models.Stack `json:"return"`
		models.APIResponse
	}

	// CreateResponse is the response from calling the create stack api
	CreateResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}

	// ListResponse is the response for listing stacks
	ListResponse struct {
		Return struct {
			Stacks *[]models.Stack `json:"data"`
		}
		models.APIResponse
	}

	// StartStopResponse represents the response from start/stop/restart actions
	StartStopResponse struct {
		Return struct {
			JobID string `json:"job_id"`
		} `json:"return"`
		models.APIResponse
	}
)
