package stack

import "github.com/sitehostnz/gosh/pkg/models"

type (
	GetResponse struct {
		Stack *models.Stack `json:"return"`
		models.APIResponse
	}

	ListResponse struct {
		Return struct {
			Stacks *[]models.Stack `json:"data"`
		}
		models.APIResponse
	}
)
