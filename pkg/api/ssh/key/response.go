package sshkey

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the listing of SSHKeys.
	ListResponse struct {
		Return struct {
			SSHKeys *[]models.SSHKey `json:"data"`
		}
		models.APIResponse
	}
)
