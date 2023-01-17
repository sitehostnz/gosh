package stack

type (
	ListRequest struct {
		ServerName string `json:"server_name"`
	}

	GetRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}

	StopStartRequest struct {
		ServerName string `json:"server_name"`
		Name       string `json:"name"`
	}
)
