package models

type (
	// SSHKey represents an SSH key in the SiteHost Api
	SSHKey struct {
		ID          string `json:"id"`
		ClientId    string `json:"client_id"`
		Label       string `json:"label"`
		Content     string `json:"content"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}
)
