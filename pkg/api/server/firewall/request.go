package firewall

type (
	// GetRequest represents a request to get a server's firewall configuration.
	GetRequest struct {
		ClientID   string `json:"client_id"`
		ServerName string `json:"server"`
	}

	// UpdateRequest represents a request to update a server's firewall.
	UpdateRequest struct {
		ClientID       string   `json:"client_id"`
		ServerName     string   `json:"server"`
		SecurityGroups []string `json:"groups"`
	}
)
