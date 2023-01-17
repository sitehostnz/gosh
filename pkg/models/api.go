package models

// APIInfo represents details about the site host api and modules the user has access to
type APIInfo struct {
	ClientID  string    `json:"client_id"`
	Modules   *[]string `json:"modules"`
	ContactID string    `json:"contact_id"`
	Roles     *[]string `json:"roles"`
}
