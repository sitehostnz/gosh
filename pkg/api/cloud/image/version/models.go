package version

import "github.com/sitehostnz/gosh/pkg/api"

type (
	// Client is a service for the /cloud/image/version endpoints.
	Client struct {
		client *api.Client
	}

	// Version describes one build of a custom image.
	Version struct {
		ID             string `json:"id"`
		ClientID       string `json:"client_id"`
		ImageID        string `json:"image_id"`
		Version        string `json:"version"`
		Labels         string `json:"labels"`
		DateAdded      string `json:"date_added"`
		DateUpdated    string `json:"date_updated"`
		IsMissing      string `json:"is_missing"`
		ForceConfig    string `json:"force_config"`
		BuildID        string `json:"build_id"`
		BuildStatus    string `json:"build_status"`
		Pending        any    `json:"pending"`
		Code           string `json:"code"`
		ContainerCount int    `json:"container_count"`
	}
)

// New initialises a version Client.
func New(c *api.Client) *Client {
	return &Client{client: c}
}
