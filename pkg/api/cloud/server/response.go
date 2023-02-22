package server

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents a server response for listing stack servers.
	ListResponse struct {
		CloudServers []models.CloudServer `json:"return"`
		models.APIResponse
	}
)
