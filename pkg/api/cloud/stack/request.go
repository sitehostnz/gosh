package stack

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListRequest represents a listing request for stacks on a server.
	ListRequest struct {
		ServerName string `json:"server_name"`
	}

	// GetRequest represents a request to get a specific stack.
	GetRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	// AddRequest represents the construction / setup of a new cloud stack.
	AddRequest struct {
		ServerName           string `json:"server_name"`
		Name                 string `json:"name"`
		Label                string `json:"label"`
		EnableSSL            int    `json:"enable_ssl"`
		DockerCompose        string `json:"docker_compose"`
		EnvironmentVariables []models.EnvironmentVariable
	}

	// StopStartRestartRequest is a request to start, restart or stop a cloud stack/container.
	StopStartRestartRequest struct {
		ServerName string   `json:"server_name"`
		Name       string   `json:"name"`
		Containers []string `json:"containers"`
	}

	// UpdateRequest modifies an existing stack. ServerName and
	// Name are required (they identify the stack). Label,
	// DockerCompose, and EnvironmentVariables are sent when
	// non-empty; leave them at their zero value to skip updating
	// that property. EnableSSL is always sent (0 or 1).
	UpdateRequest struct {
		ServerName           string `json:"server_name"`
		Name                 string `json:"name"`
		Label                string `json:"label"`
		EnableSSL            int    `json:"enable_ssl"`
		DockerCompose        string `json:"docker_compose"`
		EnvironmentVariables []models.EnvironmentVariable
	}

	// DeleteRequest removes an existing stack. Both fields are
	// required.
	DeleteRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	// CopyRequest duplicates a stack onto a destination server.
	// All four fields are required. SourceServer and
	// DestinationServer may be the same when copying within a
	// single server. Label is the new stack's label; the new
	// stack's name is the same as the source's (the API does not
	// expose a destination-name override on copy).
	CopyRequest struct {
		SourceServer      string `json:"source_server"`
		Name              string `json:"name"`
		DestinationServer string `json:"destination_server"`
		Label             string `json:"label"`
	}

	// OverwriteRequest replaces a destination stack's contents
	// with a source stack's. All four fields are required.
	// DestinationStack identifies the existing target by name on
	// DestinationServer; SourceServer + Name identify the source.
	OverwriteRequest struct {
		SourceServer      string `json:"source_server"`
		Name              string `json:"name"`
		DestinationServer string `json:"destination_server"`
		DestinationStack  string `json:"destination_stack"`
	}

	// BackupRequest creates a backup of a stack. Both fields are
	// required.
	BackupRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	// PurgeCacheRequest clears cached content from a stack. Both
	// fields are required.
	PurgeCacheRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}
)
