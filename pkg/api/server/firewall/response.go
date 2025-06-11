package firewall

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse represents the result of a request to retrieve a server's firewall configuration.
	GetResponse struct {
		Return []struct {
			Group     string `json:"group"`
			Pos       string `json:"pos"`
			Interface string `json:"interface"`
			IsRA      bool   `json:"is_ra"`
		} `json:"return"`
		Message string `json:"msg"`
		Status  bool   `json:"status"`
	}

	// UpdateResponse represents the result of a request to update a server's firewall call.
	UpdateResponse struct {
		Return struct {
			Job struct {
				Type string `json:"type"`
				Id   int    `json:"id"`
			} `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
