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

	GetRequest struct {
		MySQLHost  string `json:"mysql_host"`
		ServerName string `json:"server_name"`
		Username   string `json:"username"`
	}

	UpdateRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}

	AddRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	DeleteRequest struct {
		MySQLHost  string `json:"mysql_host"`
		ServerName string `json:"server_name"`
		Username   string `json:"username"`
	}
)
