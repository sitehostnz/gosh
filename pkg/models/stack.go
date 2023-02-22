package models

type (
	// Container represents a container inside a stack.
	Container struct {
		Name        string      `json:"name"`
		ContainerID string      `json:"container_id"`
		State       string      `json:"state"`
		Size        string      `json:"size"`
		DateCreated string      `json:"date_created"`
		SslEnabled  bool        `json:"ssl_enabled"`
		IsMissing   interface{} `json:"is_missing"`
		Pending     interface{} `json:"pending"`
	}

	// Stack represents a cloud stack and it's configuration.
	Stack struct {
		ClientID    string      `json:"client_id"`
		ServerID    string      `json:"server_id"`
		Server      string      `json:"server_name"`
		ServerLabel string      `json:"server_label"`
		Name        string      `json:"name"`
		Label       string      `json:"label"`
		DockerFile  string      `json:"docker_file"`
		IPAddress   string      `json:"ip_addr_server"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
		Containers  []Container `json:"containers"`
		ServerOwner bool        `json:"server_owner"`
		Pending     interface{} `json:"pending"`
		IsMissing   interface{} `json:"is_missing"`
	}

	// EnvironmentVariable is a stack environment variable key-pair.
	EnvironmentVariable struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
)
