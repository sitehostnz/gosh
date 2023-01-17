package models

type (
	// Domain base representation of a domain.
	Domain struct {
		Name       string `json:"name"`
		ClientID   string `json:"client_id"`
		TemplateID string `json:"template_id"`
	}
)
