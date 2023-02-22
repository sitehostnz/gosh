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

	// AddRequest represents the construction / setup of a new cloud stack.
	AddRequest struct {
		ServerName    string `json:"server_name"`
		Name          string `json:"name"`
		Label         string `json:"label"`
		EnableSSL     int    `json:"enable_ssl"`
		DockerCompose string `json:"docker_compose"`
		Environments  []models.EnvironmentVariable
	}

	// StopStartRestartRequest is a request to start, restart or stop a cloud stack/container.
	StopStartRestartRequest struct {
		ServerName string   `json:"server_name"`
		Name       string   `json:"name"`
		Containers []string `json:"containers"`
	}
)
