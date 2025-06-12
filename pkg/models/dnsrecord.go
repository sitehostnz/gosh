package models

type (
	// DNSRecord A dns record from a zone.
	DNSRecord struct {
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
