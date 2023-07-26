package grant

type (
	// AddRequest is the request object for the `/cloud/db/grant/add.json` API endpoint.
	AddRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	// UpdateRequest is the request object for the `/cloud/db/user/update.json` API endpoint.
	UpdateRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	// DeleteRequest is the request object for the `/cloud/db/user/delete.json` API endpoint.
	DeleteRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Username   string `json:"username"`
		Database   string `json:"database"`
	}
)
