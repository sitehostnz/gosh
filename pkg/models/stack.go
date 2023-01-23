package models

type Container struct {
	Name        string `json:"name"`
	ContainerId string `json:"container_id"`
	State       string `json:"state"`
	Size        string `json:"size"`
	DateCreated string `json:"date_created"`

	SslEnabled bool `json:"ssl_enabled"`

	IsMissing interface{} `json:"is_missing"`
	Pending   interface{} `json:"pending"`
}

type Stack struct {
	ClientID    string `json:"client_id"`
	ServerID    string `json:"server_id"`
	Server      string `json:"server_name"`
	ServerLabel string `json:"server_label"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	DockerFile  string `json:"docker_file"`
	ServerOwner bool   `json:"server_owner"`
	IpAddress   string `json:"ip_addr_server"`
	DateAdded   string `json:"date_added"`
	DateUpdated string `json:"date_updated"`

	Containers *[]Container `json:"containers"`

	Pending   interface{} `json:"pending"`
	IsMissing interface{} `json:"is_missing"`
}

type EnvironmentVariable struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
