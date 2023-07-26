package key

type (
	// GetRequest represents a request to get a specific SSH key.
	GetRequest struct {
		ID string `json:"key_id"`
	}
	// AddRequest is for adding a new public key with the given label.
	AddRequest struct {
		Label   string `json:"label"`
		Content string `json:"content"`
	}
	// UpdateRequest updates the key specified by the given id.
	UpdateRequest struct {
		ID      string `json:"key_id"`
		Label   string `json:"label"`
		Content string `json:"content"`
	}
	// RemoveRequest does what it says on the can, it removes the key with the given id.
	RemoveRequest struct {
		ID string `json:"key_id"`
	}
)
