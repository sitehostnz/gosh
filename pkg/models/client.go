package models

import "net/url"

// ClientBase includes the base variables for the client struct.
type ClientBase struct {
	// Base URL for API request. Always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the SiteHost API.
	UserAgent string

	// API Credentials
	APIKey   string
	ClientID string
}
