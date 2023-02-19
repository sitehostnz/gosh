package stack

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListRequest struct {
		ServerName string `json:"server_name"`
	}

	GetRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	CreateRequest struct {
		ServerName    string `json:"server_name"`
		Name          string `json:"name"`
		Label         string `json:"label"`
		EnableSSL     int    `json:"enable_ssl"`
		DockerCompose string `json:"docker_compose"`
		Environments  *map[string]models.EnvironmentVariable
	}

	StopStartRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}
)
