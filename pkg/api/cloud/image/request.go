package image

type (

	// GetRequest represents a request to get a specific image from the /cloud/image/get.json endpoint.
	GetRequest struct {
		Code string `json:"code"`
	}
)
