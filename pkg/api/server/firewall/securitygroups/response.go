package securitygroups

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// GetResponse represents a result of a get security group call.
	GetResponse struct {
		Return struct {
			Label       string `json:"label"`
			Name        string `json:"name"`
			Pending     string `json:"pending"`
			IsMissing   bool   `json:"is_missing"`
			DateAdded   string `json:"date_added"`
			DateUpdated string `json:"date_updated"`
			Servers     []struct {
				Name  string `json:"name"`
				Label string `json:"label"`
			} `json:"servers"`
			Rules struct {
				In  []Rule `json:"in"`
				Out []Rule `json:"out"`
			} `json:"rules"`
		} `json:"return"`
		models.APIResponse
	}

	// Rule represents a security group firewall rule.
	Rule struct {
		Pos        int    `json:"pos"`
		Dir        string `json:"dir"`
		Enabled    bool   `json:"enabled"`
		Action     string `json:"action"`
		Protocol   string `json:"protocol"`
		SrcIP      string `json:"src_ip"`
		SrcIPType  string `json:"src_ip_type"`
		SrcPort    string `json:"src_port"`
		DestIP     string `json:"dest_ip"`
		DestIPType string `json:"dest_ip_type"`
		DestPort   int    `json:"dest_port"`
	}

	// AddResponse represents a result of creating a security group.
	AddResponse struct {
		Return struct {
			Name string `json:"name"`
		} `json:"return"`
		models.APIResponse
	}

	// ListResponse represents a result of listing all security groups.
	ListResponse struct {
		Return struct {
			models.Pagination
			Data []struct {
				Name        string   `json:"name"`
				Label       string   `json:"label"`
				Version     int      `json:"version"`
				Servers     []string `json:"servers"`
				DateUpdated string   `json:"date_updated"`
				Pending     string   `json:"pending"`
				IsMissing   bool     `json:"is_missing"`
			} `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse represents a result of an update security group call.
	UpdateResponse struct {
		Return struct {
			Type  string `json:"type"`
			JobID string `json:"id"`
		} `json:"return"`
		models.APIResponse
	}
)
