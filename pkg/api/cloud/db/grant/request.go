package grant

type (
	AddRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	UpdateRequest struct {
		ServerName string   `json:"server_name"`
		MySQLHost  string   `json:"mysql_host"`
		Username   string   `json:"username"`
		Database   string   `json:"database"`
		Grants     []string `json:"grants"`
	}

	DeleteRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Username   string `json:"username"`
		Database   string `json:"database"`
	}
)
