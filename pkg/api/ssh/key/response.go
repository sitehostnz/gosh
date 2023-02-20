package sshkey

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListResponse struct {
		Return struct {
			SSHKeys *[]models.SSHKey `json:"data"`
		}
		models.APIResponse
	}
)
