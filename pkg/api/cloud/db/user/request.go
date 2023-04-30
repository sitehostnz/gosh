package user

type (
	// AddRequest adds/creates a new database user.
	AddRequest struct {
		ServerName string   `url:"server_name"`
		MySQLHost  string   `url:"mysql_host"`
		Username   string   `url:"username"`
		Password   string   `url:"password"`
		Database   string   `url:"database"`
		Grants     []string `url:"grants"`
	}

	// DeleteRequest a request to delete the database user.
	DeleteRequest struct {
		ServerName string `url:"server_name"`
		MySQLHost  string `url:"mysql_host"`
		Username   string `url:"username"`
	}

	// ListOptions are options for filtering/listing database users.
	ListOptions struct {
		ServerName string `url:"filters[server_name],omitempty"`
		MySQLHost  string `url:"filters[mysql_host],omitempty"`
		Username   string `url:"filters[username],omitempty"`
		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}

	// UpdateRequest is a request to the update endpoint, it only changes the backup container.
	UpdateRequest struct {
		ServerName string `url:"server_name"`
		MySQLHost  string `url:"mysql_host"`
		Database   string `url:"database"`
		Container  string `url:"params[container]'"`
	}

	// GetRequest is for getting a single database user.
	GetRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Username   string `json:"username"`
	}
)
