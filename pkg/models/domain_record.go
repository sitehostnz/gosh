package models

type (
	DomainRecord struct {
		Id         string `json:"id"`
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
