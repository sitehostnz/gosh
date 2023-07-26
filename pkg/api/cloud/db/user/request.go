package user

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListOptions are options for filtering/listing databases.
	ListOptions struct {
		ServerName string `url:"filters[server_name],omitempty"`
		Username   string `url:"filters[username],omitempty"`
		MySQLHost  string `url:"filters[mysql_host],omitempty"`
		models.Filtering
	}

	// GetRequest represents a request to get the specified database user on the given server/host.
	GetRequest struct {
		MySQLHost  string `json:"mysql_host"`
		ServerName string `json:"server_name"`
		Username   string `json:"username"`
	}

	// UpdateRequest represents a request to update a database user. It sets the password.
	UpdateRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}

	// AddRequest represents a request to add a database user, with an initial set of grants.
	AddRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	// DeleteRequest represents a request to delete a database user.
	DeleteRequest struct {
		MySQLHost  string `json:"mysql_host"`
		ServerName string `json:"server_name"`
		Username   string `json:"username"`
	}
)
