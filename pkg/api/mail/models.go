package mail

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

type (
	// Client is a Service to work with the SiteHost Mail API.
	// defaultServerName, when non-empty, is used as the server_name
	// for any operation whose request didn't provide one — see
	// NewForServer.
	Client struct {
		client            *api.Client
		defaultServerName string
	}
)

// New returns a Mail client with no captured default server.
// Every operation must specify the ServerName explicitly in its
// request options.
func New(c *api.Client) *Client {
	return &Client{client: c}
}

// NewForServer returns a Mail client that captures serverName as
// the default for every operation. Per-call ServerName values in
// request options override the captured default; an empty value
// in the request options falls back to the default.
//
// Use this when a consumer only operates against one mail
// service. For multi-service consumers, prefer New and pass
// ServerName explicitly on each call.
func NewForServer(c *api.Client, serverName string) *Client {
	return &Client{client: c, defaultServerName: serverName}
}

// resolveServerName picks the per-call value when set, otherwise
// the captured default. Returns empty if neither is set; callers
// are expected to validate before issuing the request.
func (s *Client) resolveServerName(perCall string) string {
	if perCall != "" {
		return perCall
	}
	return s.defaultServerName
}
