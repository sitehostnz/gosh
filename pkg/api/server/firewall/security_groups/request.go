package security_groups

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// AddRequest represents a request to create a security group.
	AddRequest struct {
		ClientID string `json:"client_id"`
		Label    string `json:"label"`
	}

	// DeleteRequest represents a request to remove a security group.
	DeleteRequest struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}

	// GetRequest represents a request to get details of a security group.
	GetRequest struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}

	// ListAllRequest represents options for filtering security groups.
	ListAllRequest struct {
		Label string `url:"filters[label],omitempty"`
		models.Filtering
	}

	// UpdateRequest represents a request to update a security group.
	UpdateRequest struct {
		ClientID string        `json:"client_id"`
		Name     string        `json:"name"`
		Params   ParamsOptions `json:"params"`
	}

	// ParamsOptions represents options for updating a security group.
	ParamsOptions struct {
		Label    string    `json:"label"`
		RulesIn  []RuleIn  `json:"rules_in"`
		RulesOut []RuleOut `json:"rules_out"`
	}

	// RuleIn represents a request to update an inbound rule for a security group.
	RuleIn struct {
		Enabled         bool   `json:"enabled"`
		Action          string `json:"action"`
		Protocol        string `json:"protocol"`
		SourceIP        string `json:"src_ip"`
		DestinationPort string `json:"dest_port"`
	}

	// RuleOut represents a request to update an outbound rule for a security group.
	RuleOut struct {
		Enabled         bool   `json:"enabled"`
		Action          string `json:"action"`
		Protocol        string `json:"protocol"`
		DestinationIP   string `json:"dest_ip"`
		DestinationPort string `json:"dest_port"`
	}
)
