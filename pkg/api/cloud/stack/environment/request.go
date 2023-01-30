package environment

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetRequest struct {
		ServerName string `json:"server"`
		Project    string `json:"project"`
		Service    string `json:"service"`
	}
	UpdateRequest struct {
		ServerName           string `json:"server"`
		Project              string `json:"project"`
		Service              string `json:"service"`
		EnvironmentVariables *[]models.EnvironmentVariable
	}
)
