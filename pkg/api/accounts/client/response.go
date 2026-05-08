package client

import "github.com/sitehostnz/gosh/pkg/models"

// ListSubAccountsResponse is the paginated return from
// /accounts/client/list_sub_accounts.json.
type ListSubAccountsResponse struct {
	Return struct {
		models.Pagination
		Accounts []SubAccount `json:"data"`
	} `json:"return"`
	models.APIResponse
}
