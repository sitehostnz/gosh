package securitygroups

type (
	// AddRequest represents a request to create a security group.
	AddRequest struct {
		ClientID string `json:"client_id"`
		Label    string `json:"label"`
	}
)

type (
	// DeleteRequest represents a request to remove a security group.
	DeleteRequest struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}
)

type (
	// GetRequest represents a request to get details of a security group.
	GetRequest struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}
)

type (
	// UpdateRequest represents a request to update a security group.
	UpdateRequest struct {
		ClientID string        `json:"client_id"`
		Name     string        `json:"name"`
		Params   ParamsOptions `json:"params"`
	}

	ParamsOptions struct {
		Label    string    `json:"label"`
		RulesIn  []RuleIn  `json:"rules_in"`
		RulesOut []RuleOut `json:"rules_out"`
	}

	RuleIn struct {
		Enabled         bool   `json:"enabled"`
		Action          string `json:"action"`
		Protocol        string `json:"protocol"`
		SourceIP        string `json:"src_ip"`
		DestinationPort string `json:"dest_port"`
	}

	RuleOut struct {
		Enabled         bool   `json:"enabled"`
		Action          string `json:"action"`
		Protocol        string `json:"protocol"`
		DestinationIP   string `json:"dest_ip"`
		DestinationPort string `json:"dest_port"`
	}
)
