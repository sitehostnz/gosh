package db

type (
	// AddRequest adds/creates a new database on the given cloud server.
	AddRequest struct {
		ServerName string `url:"server_name"`
		MySQLHost  string `url:"mysql_host"`
		Database   string `url:"database"`
		Container  string `url:"container"`
	}

	// DeleteRequest a request to delete the database.
	DeleteRequest struct {
		ServerName string `url:"server_name"`
		MySQLHost  string `url:"mysql_host"`
		Database   string `url:"database"`
	}

	// UpdateRequest is a request to the update endpoint, it only changes the backup container.
	UpdateRequest struct {
		ServerName string `url:"server_name"`
		MySQLHost  string `url:"mysql_host"`
		Database   string `url:"database"`
		Container  string `url:"params[container]'"`
	}

	// ListOptions are options for filtering/listing databases.
	ListOptions struct {
		ServerName string `url:"filters[server_name],omitempty"`
		MySQLHost  string `url:"filters[mysql_host],omitempty"`
		Database   string `url:"filters[db_name],omitempty"`

		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}

	// GetRequest is for getting a single database.
	GetRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Database   string `json:"database"`
	}
)
