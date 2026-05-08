package letsencrypt

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

type (
	// Client is a Service for Cloud Container Let's Encrypt
	// management.
	Client struct {
		client *api.Client
	}
)

// New is an initialisation function.
func New(c *api.Client) *Client {
	return &Client{client: c}
}
