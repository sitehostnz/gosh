package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListDomainsResponse struct {
		Return struct {
			models.Pagination
			Domains []models.Domain `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	ListContactsResponse struct {
		DomainContacts []models.DomainContact `json:"return"`

		models.APIResponse
	}
)
