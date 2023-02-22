package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetRequest represents a request to get a stacks environment.
	GetRequest struct {
		ServerName string `json:"server"`
		Project    string `json:"project"`
		Service    string `json:"service"`
	}

	// UpdateRequest represents a call to update the environment variables of a stack.
	UpdateRequest struct {
		ServerName           string `json:"server"`
		Project              string `json:"project"`
		Service              string `json:"service"`
		EnvironmentVariables []models.EnvironmentVariable
	}
)
