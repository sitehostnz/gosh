package user

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents the return from the /cloud/ssh/user/list_all.json endpoint.
	ListResponse struct {
		Return struct {
			models.Pagination
			Users []models.User `json:"data"`
		}
		models.APIResponse
	}

	// GetResponse represents the return from the /cloud/ssh/user/get.json endpoint.
	GetResponse struct {
		Return models.User `json:"return"`
		models.APIResponse
	}

	// AddResponse is the return for adding an ssh/sftp user.
	AddResponse struct {
		Return struct {
			Job models.JobResponse
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse is the return for updating an ssh/sftp user.
	UpdateResponse struct {
		Return struct {
			Job models.JobResponse
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse is the return for deleting a ssh/sftp user.
	DeleteResponse struct {
		Return struct {
			Job models.JobResponse
		} `json:"return"`
		models.APIResponse
	}
)
