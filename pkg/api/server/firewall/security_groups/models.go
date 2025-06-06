package security_groups

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

// Client is a Service to work with API Jobs.
type Client struct {
	client *api.Client
}

// New is used to instantiate the Client struct.
func New(c *api.Client) *Client {
	return &Client{
		client: c,
	}
}
