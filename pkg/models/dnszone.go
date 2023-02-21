package models

type (
	// DNSZone base representation of a domain.
	DNSZone struct {
		Name       string `json:"name"`
		ClientID   string `json:"client_id"`
		TemplateID string `json:"template_id"`
		Pending    string `json:"pending"`
	}
)
