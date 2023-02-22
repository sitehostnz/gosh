package stackserver

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents a server response for listing stack servers.
	ListResponse struct {
		StackServers *[]models.StackServer `json:"return"`
		models.APIResponse
	}
)
