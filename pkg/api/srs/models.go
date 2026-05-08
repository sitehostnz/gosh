package srs

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

type (
	// Client is a Service to work with the SiteHost SRS API.
	Client struct {
		client *api.Client
	}
)

// New is an initialisation function.
func New(c *api.Client) *Client {
	return &Client{
		client: c,
	}
}
