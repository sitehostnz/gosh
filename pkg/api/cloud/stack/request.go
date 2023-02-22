package stack

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListRequest represents a listing request for stacks on a server.
	ListRequest struct {
		ServerName string `json:"server_name"`
	}

	// GetRequest represents a request to get a specific stack.
	GetRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	// CreateRequest represents the construction / setup of a new cloud stack.
	CreateRequest struct {
		ServerName    string `json:"server_name"`
		Name          string `json:"name"`
		Label         string `json:"label"`
		EnableSSL     int    `json:"enable_ssl"`
		DockerCompose string `json:"docker_compose"`
		Environments  *map[string]models.EnvironmentVariable
	}

	// StopStartRequest is a request to start or stop a server.
	StopStartRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}
)
