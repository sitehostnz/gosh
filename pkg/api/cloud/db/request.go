package db

type (
	// ListOptions are options for filtering/listing databases
	ListOptions struct {
		ServerName string `url:"filters[server_name],omitempty"`
		MySQLHost  string `url:"filters[mysql_host],omitempty"`
		Database   string `url:"filters[db_name],omitempty"`

		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}
	// GetRequest is for getting a single database
	GetRequest struct {
		ServerName string `json:"server_name"`
		MySQLHost  string `json:"mysql_host"`
		Database   string `json:"database"`
	}
)
