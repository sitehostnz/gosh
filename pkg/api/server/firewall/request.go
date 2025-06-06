package firewall

type (
	// UpdateRequest represents a request to update a server's firewall.
	UpdateRequest struct {
		ServerName     string   `json:"server"`
		SecurityGroups []string `json:"groups"`
	}
)
