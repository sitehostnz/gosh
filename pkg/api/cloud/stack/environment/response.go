package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse is the response contains the results of a call to get an environment.
	// It contains the environment details and variables.
	GetResponse struct {
		EnvironmentVariables []models.EnvironmentVariable `json:"return"`
		models.APIResponse
	}

	// UpdateResponse is the result of adding or updating variables in a stack environment.
	UpdateResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse is the result of deleting variables from a stack environment.
	DeleteResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
