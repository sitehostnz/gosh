package stackserver

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListResponse struct {
		StackServers *[]models.StackServer `json:"return"`
		models.APIResponse
	}
)
