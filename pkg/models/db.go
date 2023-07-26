package models

type (
	// Grant represents database grants.
	Grant struct {
		DBName      string      `json:"db_name"`
		Username    string      `json:"username"`
		Host        string      `json:"host"`
		MySQLHost   string      `json:"mysql_host"`
		Grants      []string    `json:"grants"`
		Pending     interface{} `json:"pending"`
		IsMissing   string      `json:"is_missing"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
	}

	// Database represents a database on a cloud server. Some properties may or may not be populated depending on how we call.
	Database struct {
		ID          string      `json:"id"`
		DBName      string      `json:"db_name"`
		MySQLHost   string      `json:"mysql_host"`
		Size        interface{} `json:"size"`
		ClientID    string      `json:"client_id"`
		ServerID    string      `json:"server_id"`
		Pending     string      `json:"pending"`
		IsMissing   string      `json:"is_missing"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
		ServerName  string      `json:"server_name"`
		ServerLabel string      `json:"server_label"`
		ServerIP    string      `json:"server_ip"`
		ServerOwner bool        `json:"server_owner"`
		Container   string      `json:"container"`
		Grants      []Grant     `json:"grants"`
	}

	// DatabaseUser represents a database user on a cloud server.
	DatabaseUser struct {
		Username    string      `json:"username"`
		Password    string      `json:"password"`
		MysqlHost   string      `json:"mysql_host"`
		ServerID    string      `json:"server_id"`
		ClientID    string      `json:"client_id"`
		Pending     interface{} `json:"pending"`
		IsMissing   string      `json:"is_missing"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
		ServerName  string      `json:"server_name"`
		ServerLabel string      `json:"server_label"`
		Grants      []Grant     `json:"grants"`
	}
)
