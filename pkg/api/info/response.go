package info

// GetResponse represents a result of a get API Info call.
type GetResponse struct {
	Return struct {
		ClientID    string      `json:"client_id"`
		ContactID   interface{} `json:"contact_id"`
		Modules     []string    `json:"modules"`
		Roles       []string    `json:"roles"`
		Environment struct {
			GitBranch string `json:"git_branch"`
			Type      string `json:"type"`
		} `json:"environment"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}
