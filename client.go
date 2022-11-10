// Package gosh provides the functions to work with SiteHost API.
package gosh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "https://api.staging.sitehost.nz"
	defaultVersion = "1.1"
	userAgent      = "gosh"

	defaultMediaType = "application/x-www-form-urlencoded"
)

// ErrorResponse reports the error caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Message  string         `json:"msg"`
	Status   bool           `json:"status"`
}

// Error returns a ErrorResponse message.
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// Client is a wrapper around the http client to manages communication with SiteHost API V1.1.
type Client struct {
	client *http.Client

	// Base URL for API request. Always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the SiteHost API.
	UserAgent string

	// API Credentials
	APIKey   string
	ClientID string

	common service

	// Services used for talking to different par of the SiteHost API.
	Servers *ServersService
	Jobs    *JobsService
}

type service struct {
	client *Client
}

// ClientOpt function parameters to configure a Client.
type ClientOpt func(*Client) error

// New returns a new SiteHost API client instance.
func New(apiKey, clientID string, opts ...ClientOpt) (*Client, error) {
	c := NewClient(apiKey, clientID)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// NewClient factory to create new Client struct.
func NewClient(apiKey, clientID string) *Client {
	baseURL, _ := url.Parse(fmt.Sprintf("%s/%s/", defaultBaseURL, defaultVersion))

	c := &Client{
		client:    &http.Client{},
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ClientID:  clientID,
		UserAgent: userAgent,
	}
	c.common.client = c
	c.Servers = (*ServersService)(&c.common)
	c.Jobs = (*JobsService)(&c.common)

	return c
}

// SetBaseURL change the default BaseURL.
func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}
		c.BaseURL = u
		return nil
	}
}

// NewRequest creates an SiteHost API Request.
func (c *Client) NewRequest(method, uri string, body string) (*http.Request, error) {
	u, err := c.BaseURL.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Check if APIKey or Client ID are empty.
	if c.APIKey == "" || c.ClientID == "" {
		return nil, fmt.Errorf("API Key and Client ID must be different to empty")
	}

	values := u.Query()
	values.Add("apikey", c.APIKey)
	values.Add("client_id", c.ClientID)
	u.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", defaultMediaType)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// Do send an API Request and returns the response.
//
// The API response is checked  to see if it was a successful call.
// A successful call is then checked to see if we have a Status true.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error when closing", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check if the Status message is true.
	if err = CheckResponse(resp, body); err != nil {
		return err
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return err
		}
	}

	return nil
}

// CheckResponse checks the API response for errors, and returns them if present.
//
// A response is considered an error if it has a status code outside the 200 range or if the Status is false.
func CheckResponse(r *http.Response, data []byte) error {
	errorResponse := &ErrorResponse{Response: r}
	if err := json.Unmarshal(data, errorResponse); err == nil {
		if errorResponse.Status {
			return nil
		}
	}

	return errorResponse
}
