package models

type ApiInfo struct {
	ClientID  string    `json:"client_id"`
	Modules   *[]string `json:"modules"`
	ContactID string    `json:"contact_id"`
	Roles     *[]string `json:"roles"`
}
