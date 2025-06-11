// Package api provides the functions to work with SiteHost API.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sitehostnz/gosh/pkg/models"
)

const (
	defaultBaseURL   = "https://api.sitehost.nz"
	defaultVersion   = "1.5"
	userAgent        = "gosh"
	defaultMediaType = "application/x-www-form-urlencoded"
)

type (
	// Client is a wrapper around the http client to manages communication with SiteHost API.
	Client struct {
		client *http.Client
		models.ClientBase
	}

	// ClientOpt function parameters to configure a Client.
	ClientOpt func(*Client) error
)

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

// Do sends an API Request and returns the response.
//
// The API response is checked  to see if it was a successful call.
// A successful call is then checked to see if we have a Status true.
func (c *Client) Do(_ context.Context, req *http.Request, v interface{}) error {
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
	if err := CheckResponse(resp, body); err != nil {
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
	errorResponse := &models.ErrorResponse{Response: r}
	if err := json.Unmarshal(data, errorResponse); err == nil {
		if errorResponse.Status {
			return nil
		}
	}

	return errorResponse
}

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
		client: &http.Client{},
		ClientBase: models.ClientBase{
			BaseURL:   baseURL,
			APIKey:    apiKey,
			ClientID:  clientID,
			UserAgent: userAgent,
		},
	}

	return c
}

// SetBaseURL change the default BaseURL (Include the API version).
func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(fmt.Sprintf("%s/", bu))
		if err != nil {
			return err
		}
		c.BaseURL = u
		return nil
	}
}
