package image

type (
	// GetRequest is the request to get a image by code.
	GetRequest struct {
		Code string `json:"code"`
	}
)
