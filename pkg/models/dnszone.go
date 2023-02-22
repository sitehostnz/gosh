package models

type (
	// DNSZone represent a DNS domain.
	DNSZone struct {
		Name       string `json:"name"`
		ClientID   string `json:"client_id"`
		TemplateID string `json:"template_id"`
		Pending    string `json:"pending"`
	}
)
