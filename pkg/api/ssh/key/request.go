package key

type (
	// GetRequest represents a request to get a specific SSH Key.
	GetRequest struct {
		ID string `json:"key_id"`
	}

	// CreateRequest represents a request to create an SSH Key.
	CreateRequest struct {
		Label             string `json:"label"`
		Content           string `json:"content"`
		CustomImageAccess string `json:"custom_image_access"`
	}

	// DeleteRequest represents a request to delete a specific SSH Key.
	DeleteRequest struct {
		ID string `json:"key_id"`
	}

	// UpdateRequest represents a request to update a specific SSH Key.
	UpdateRequest struct {
		ID      string `json:"key_id"`
		Label   string `json:"label"`
		Content string `json:"content"`
		CustomImageAccess string `json:"custom_image_access"`
	}
}
