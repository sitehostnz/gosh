package gosh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type Client struct {
	client *http.Client

	// Base URL for API request. Always be specified with a trailing slash
	BaseURL *url.URL

	// User agent used when communicating with the SiteHost API
	UserAgent string

	APIKey   string
	ClientID string

	common service

	// Services used for talking to different par of the SiteHost API
	Servers *ServersService
}

type service struct {
	client *Client
}

type ClientOpt func(*Client) error

func New(apiKey, clientID string, opts ...ClientOpt) (*Client, error) {
	c := NewClient(apiKey, clientID)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

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

	return c
}

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

func (c *Client) NewRequest(method, uri string, body string) (*http.Request, error) {
	u, err := c.BaseURL.Parse(uri)
	if err != nil {
		return nil, err
	}

	// TODO: Add apikey and client_id in query string
	if c.APIKey == "" || c.ClientID == "" {
		return nil, fmt.Errorf("API Key and Client ID must be different to empty")
	}

	values := u.Query()
	values.Add("apikey", c.APIKey)
	values.Add("client_id", c.ClientID)
	u.RawQuery = values.Encode()

	req, err := http.NewRequest(method, u.String(), strings.NewReader(body))

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", defaultMediaType)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c Client) NewPostRequest(uri string, reader io.Reader, size int64, mediaType string) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing space, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(uri)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), reader)
	if err != nil {
		return nil, err
	}

	req.ContentLength = size

	if mediaType == "" {
		mediaType = defaultMediaType
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	//
	// TODO: CONTROL ERRORS HERE (read the body and check the status)
	//
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return err
		}
	}

	return nil
}
