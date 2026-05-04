package client

import "github.com/sitehostnz/gosh/pkg/api"

// Client wraps the api.Client for the /accounts/client endpoints.
type Client struct {
	client *api.Client
}

// New initialises an accounts.client Client.
func New(c *api.Client) *Client {
	return &Client{client: c}
}

// SubAccount describes one row returned by ListSubAccounts.
type SubAccount struct {
	ClientID       string `json:"client_id"`
	Name           string `json:"name"`
	AccountBalance string `json:"account_balance"`
	Joined         string `json:"joined"`
	// Closed is "0000-00-00" for currently-open accounts; otherwise
	// the ISO date the account was closed.
	Closed      string `json:"closed"`
	AccountType string `json:"account_type"`
}
