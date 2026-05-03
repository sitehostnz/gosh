package redirect

import "github.com/sitehostnz/gosh/pkg/api"

// Client wraps the api.Client for the /redirect endpoints.
type Client struct {
	client *api.Client
}

// New initialises a redirect Client.
func New(c *api.Client) *Client {
	return &Client{client: c}
}

// Rule describes one redirect: where the source URL points to and
// the HTTP status code (typically 301 or 302).
type Rule struct {
	Destination string `json:"destination"`
	Type        int    `json:"type"`
}
