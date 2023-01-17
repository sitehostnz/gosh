package models

type (
	// DomainRecord A dns record from a zone.
	DomainRecord struct {
		ID         string `json:"id"`
		ClientID   string `json:"client_id"`
		Name       string `json:"name"`
		Domain     string `json:"domain"`
		Type       string `json:"type"`
		Content    string `json:"content"`
		TTL        string `json:"ttl"`
		Priority   string `json:"prio"`
		ChangeDate string `json:"change_date"`
		State      string `json:"state"`
	}
)
