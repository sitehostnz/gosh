package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// NameServer is a single nameserver entry.
	NameServer struct {
		Name    string `json:"name"`
		IPv4Addr string `json:"ipv4addr"`
		IPv6Addr string `json:"ipv6addr"`
	}

	// ListNameServersResponse represents the response from
	// list_name_servers.
	ListNameServersResponse struct {
		Return []NameServer `json:"return"`
		models.APIResponse
	}
)
