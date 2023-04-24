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
		ServerName           string `json:"server_name"`
		Name                 string `json:"name"`
		Label                string `json:"label"`
		EnableSSL            int    `json:"enable_ssl"`
		DockerCompose        string `json:"docker_compose"`
		EnvironmentVariables []models.EnvironmentVariable
	}

	// AddRequestWithImage represents the construction / setup of a new cloud stack.
	AddRequestWithImage struct {
		ServerName           string `json:"server_name"`
		Name                 string `json:"name"`
		Label                string `json:"label"`
		EnableSSL            int    `json:"enable_ssl"`
		ImageCode            string `json:"image_code"`
		EnvironmentVariables []models.EnvironmentVariable
	}

	// GenerateDockerComposeRequest represents the construction / setup of a docker compose.
	GenerateDockerComposeRequest struct {
		Name      string `json:"name"`
		Label     string `json:"label"`
		ImageCode string `json:"image_code"`
	}

	// BuildDockerCompose represents the construction / setup of a docker compose.
	BuildDockerCompose struct {
		Name    string   `json:"name"`
		Label   string   `json:"label"`
		Image   string   `json:"image"`
		Type    string   `json:"type"`
		Ports   []string `yaml:"ports"`
		Volumes []string `yaml:"volumes"`
	}

	// StopStartRestartRequest is a request to start, restart or stop a cloud stack/container.
	StopStartRestartRequest struct {
		ServerName string   `json:"server_name"`
		Name       string   `json:"name"`
		Containers []string `json:"containers"`
	}
)
