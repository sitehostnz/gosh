package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ContactSummary is the brief contact entry returned by
	// list_contacts and search_contacts.
	ContactSummary struct {
		ContactID      string `json:"contact_id"`
		Name           string `json:"name"`
		RegistrantName string `json:"registrant_name"`
		Email          string `json:"email"`
		PhoneCountry   string `json:"phone_cntry"`
		PhoneArea      string `json:"phone_area"`
		PhoneLocal     string `json:"phone_local"`
		PhoneExtension string `json:"phone_extension"`
		DomainCount    int    `json:"domain_count"`
	}

	// ListContactsResponse represents the response from list_contacts.
	// Note: pagination fields are at the top level (not inside Return).
	ListContactsResponse struct {
		models.Pagination
		Return []ContactSummary `json:"return"`
		models.APIResponse
	}

	// SearchContactsResponse represents the response from search_contacts.
	SearchContactsResponse struct {
		Return []ContactSummary `json:"return"`
		models.APIResponse
	}

	// ContactDetail is the full contact record returned by get_contact.
	// Fields are PascalCase JSON keys.
	ContactDetail struct {
		ContactID       int    `json:"ContactID"`
		ClientID        string `json:"ClientID"`
		Name            string `json:"Name"`
		RegistrantName  string `json:"RegistrantName"`
		Organization    string `json:"Organization"`
		PostalAddress   string `json:"PostalAddress"`
		PostalAddress2  string `json:"PostalAddress2"`
		Suburb          string `json:"Suburb"`
		PostCode        string `json:"PostCode"`
		Province        string `json:"Province"`
		City            string `json:"City"`
		Country         string `json:"Country"`
		PhoneCountry    string `json:"PhoneCountry"`
		PhoneArea       string `json:"PhoneArea"`
		PhoneLocal      string `json:"PhoneLocal"`
		PhoneExtension  string `json:"PhoneExtension"`
		MobileCountry   string `json:"MobileCountry"`
		MobileArea      string `json:"MobileArea"`
		MobileLocal     string `json:"MobileLocal"`
		MobileExtension string `json:"MobileExtension"`
		FaxCountry      string `json:"FaxCountry"`
		FaxArea         string `json:"FaxArea"`
		FaxLocal        string `json:"FaxLocal"`
	}

	// GetContactResponse represents the response from get_contact.
	GetContactResponse struct {
		Return ContactDetail `json:"return"`
		models.APIResponse
	}
)
