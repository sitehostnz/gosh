package image

type (
	// GetRequest represents a filtered request to the /cloud/stack/image/list_all.json endpoint.
	GetRequest struct {
		Code string `json:"code"`
	}
)
