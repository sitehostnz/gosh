package models

type (
	User struct {
		Username       string      `json:"username"`
		ServerID       string      `json:"server_id"`
		ClientID       string      `json:"client_id"`
		HomeDir        string      `json:"home_dir"`
		SSHKeys        []SSHKey    `json:"ssh_keys"`
		Pending        interface{} `json:"pending"`
		IsMissing      string      `json:"is_missing"`
		DateAdded      string      `json:"date_added"`
		DateUpdated    string      `json:"date_updated"`
		ServerName     string      `json:"server_name"`
		ServerLabel    string      `json:"server_label"`
		ServerOwner    bool        `json:"server_owner"`
		Image          string      `json:"image"`
		Containers     []string    `json:"containers"`
		ReadOnlyConfig bool        `json:"read_only_config"`
		IpAddr         string      `json:"ip_addr"`
	}
)
