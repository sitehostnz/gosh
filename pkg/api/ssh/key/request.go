package key

type (
	// GetRequest represents a request to get a specific SSH key.
	GetRequest struct {
		ID string `json:"key_id"`
	}
)
