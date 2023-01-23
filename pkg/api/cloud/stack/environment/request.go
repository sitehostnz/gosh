package environment

type (
	GetRequest struct {
		ServerName string `json:"server"`
		Project    string `json:"project"`
		Service    string `json:"service"`
	}
)
