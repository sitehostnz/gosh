package server

// GetRequest represents request params for get server endpoint.
type GetRequest struct {
	ServerName string `json:"name"`
}

// DeleteRequest represents a request to delete a Server.
type DeleteRequest struct {
	Name string `json:"name"`
}

// UpgradeRequest represents a request to upgrade a Server.
type UpgradeRequest struct {
	Name string `json:"name"`
	Plan string `json:"plan"`
}

// UpdateRequest represents a request to update a Server.
type UpdateRequest struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// CommitDiskChangesRequest represents request params for CommitDiskChanges server endpoint.
type CommitDiskChangesRequest struct {
	ServerName string `json:"name"`
}

// CreateRequest represents a request to create a Server.
type CreateRequest struct {
	ClientID    string        `json:"client_id"`
	Label       string        `json:"label"`
	Location    string        `json:"location"`
	ProductCode string        `json:"product_code"`
	Image       string        `json:"image"`
	Params      ParamsOptions `json:"params"`
}

// ParamsOptions represents the additional parameters in the request to create a Server.
type ParamsOptions struct {
	Name      string   `json:"name,omitempty"`
	IPv4      []string `json:"ipv4"`
	IPv6      []string `json:"ipv6,omitempty"`
	SSHKeys   []string `json:"ssh_keys,omitempty"`
	ContactID string   `json:"contact_id,omitempty"`
	Backup    string   `json:"backup,omitempty"`
	SendEmail string   `json:"send_email,omitempty"`
}
