package info

// GetResponse represents a result of a get API Info call.
type GetResponse struct {
	Return struct {
		ClientID  string   `json:"client_id"`
		ContactID string   `json:"contact_id"`
		Modules   []string `json:"modules"`
		Roles     []string `json:"roles"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}
