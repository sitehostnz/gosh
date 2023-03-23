package user

import "github.com/sitehostnz/gosh/pkg/models"

type (

	// AddRequest adds/creates a new database on the given cloud server.
	AddRequest struct {
		ServerName     string   `url:"server_name"`
		Username       string   `url:"username"`
		Password       string   `url:"password"`
		Containers     []string `url:"containers[]"`
		SSHKeys        []string `url:"ssh_keys[],omitempty"`
		ReadOnlyConfig int      `url:"read_only_config,omitempty"`
	}
	// DeleteRequest a request to delete the database.
	DeleteRequest struct {
		ServerName string `url:"server_name"`
		Username   string `url:"username"`
	}
	// GetRequest a request to delete the database.
	GetRequest struct {
		ServerName string `url:"server_name"`
		Username   string `url:"username"`
	}
	// UpdateRequest is a request to the update endpoint
	UpdateRequest struct {
		ServerName     string   `url:"server_name"`
		Username       string   `url:"username"`
		Password       string   `url:"params[password],omitempty"`
		Containers     []string `url:"params[containers][],omitempty"`
		SSHKeys        []string `url:"params[ssh_keys][],omitempty"`
		ReadOnlyConfig int      `url:"params[read_only_config],omitempty"`
	}
	// ListOptions are options for filtering/listing users.
	ListOptions struct {
		ServerName string `url:"filters[server_name],omitempty"`
		Username   string `url:"filters[username],omitempty"`
		Container  string `url:"filters[container],omitempty"`
		models.Filtering
	}
)
