package gosh

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/go-resty/resty/v2"
)

var envDebug = false

const (
	// APIHost Sitehost API URL
	APIHost = "api.sitehost.nz"
	// APIVersion Sitehost API version
	APIVersion = "1.1"
	// APIProto connect to API with http(s)
	APIProto = "https"
	// APISecondsPerPoll how frequently to poll for new Events or Status in WaitFor functions
	APISecondsPerPoll = 3
	// APIRetryMaxWaitTime maximum wait time for retries
	APIRetryMaxWaitTime = time.Duration(30) * time.Second
)

// Client is a wrapper around the Resty Client
type Client struct {
	resty             *resty.Client
	userAgent         string
	debug             bool
	resources         map[string]*Resource
	retryConditionals []RetryConditional

	millisecondsPerPoll time.Duration

	baseURL    string
	apiVersion string
	apiProto   string

	Servers *Resource
}

// SetUserAgent sets a custom user-agent for HTTP requests
func (c *Client) SetUserAgent(ua string) *Client {
	c.userAgent = ua
	c.resty.SetHeader("User-Agent", c.userAgent)

	return c
}

// R wraps resty's R method
func (c *Client) R(ctx context.Context) *resty.Request {
	return c.resty.R().
		ExpectContentType("application/json").
		SetHeader("Content-Type", "application/json").
		SetContext(ctx).
		SetError(APIError{})
}

// SetDebug sets the debug on resty's client
func (c *Client) SetDebug(debug bool) *Client {
	c.debug = debug
	c.resty.SetDebug(debug)

	return c
}

// SetBaseURL sets the base URL of the Sitehost API (https://api.sitehost.nz/1.1)
func (c *Client) SetBaseURL(baseURL string) *Client {
	baseURLPath, _ := url.Parse(baseURL)

	c.baseURL = path.Join(baseURLPath.Host, baseURLPath.Path)
	c.apiProto = baseURLPath.Scheme

	c.updateHostURL()

	return c
}

// SetAPIVersion sets the version of the API to interface with
func (c *Client) SetAPIVersion(apiVersion string) *Client {
	c.apiVersion = apiVersion

	c.updateHostURL()

	return c
}

// updateHostURL sets sitehost's API URL
func (c *Client) updateHostURL() {
	apiProto := APIProto
	baseURL := APIHost
	apiVersion := APIVersion

	if c.baseURL != "" {
		baseURL = c.baseURL
	}

	if c.apiVersion != "" {
		apiVersion = c.apiVersion
	}

	if c.apiProto != "" {
		apiProto = c.apiProto
	}

	c.resty.SetBaseURL(fmt.Sprintf("%s://%s/%s", apiProto, baseURL, apiVersion))
}

// SetToken sets the API Key token for all requests from this client
func (c *Client) SetToken(token, clientID string) *Client {
	c.resty.SetQueryParam("apikey", token)
	c.resty.SetQueryParam("client_id", clientID)

	return c
}

// SetRetries adds retry conditions for Sitehost errors and HTTP 429 error
func (c *Client) SetRetries() *Client {
	c.
		addRetryConditional(tooManyRequestsRetryCondition).
		addRetryConditional(serviceUnavailableRetryCondition).
		addRetryConditional(requestTimeoutRetryCondition).
		SetRetryMaxWaitTime(APIRetryMaxWaitTime)

	configureRetries(c)

	return c
}

// addRetryConditional add retry condition to client
func (c *Client) addRetryConditional(retryConditional RetryConditional) *Client {
	c.retryConditionals = append(c.retryConditionals, retryConditional)

	return c
}

// SetRetryMaxWaitTime sets the maximum delay before retrying a request.
func (c *Client) SetRetryMaxWaitTime(max time.Duration) *Client {
	c.resty.SetRetryMaxWaitTime(max)
	return c
}

// SetRetryWaitTime sets the default (minimum) delay before retrying a request.
func (c *Client) SetRetryWaitTime(min time.Duration) *Client {
	c.resty.SetRetryWaitTime(min)
	return c
}

// SetRetryAfter sets the callback function to be invoked with a failed request
// to determine when it should be retried.
func (c *Client) SetRetryAfter(callback RetryAfter) *Client {
	c.resty.SetRetryAfter(resty.RetryAfterFunc(callback))
	return c
}

// SetRetryCount sets the maximum retry attempts before aborting.
func (c *Client) SetRetryCount(count int) *Client {
	c.resty.SetRetryCount(count)
	return c
}

// SetPollDelay sets the number of milliseconds to wait between events or status polls.
// Affects all WaitFor* functions and retries.
func (c *Client) SetPollDelay(delay time.Duration) *Client {
	c.millisecondsPerPoll = delay
	return c
}

// Resource looks up a resource by name
func (c *Client) Resource(resourceName string) *Resource {
	selectedResource, ok := c.resources[resourceName]
	if !ok {
		log.Fatalf("Could not find resource named '%s', exiting.", resourceName)
	}

	return selectedResource
}

// NewClient factory to create new Client struct
func NewClient(hc *http.Client) (client Client) {
	if hc != nil {
		client.resty = resty.NewWithClient(hc)
	} else {
		client.resty = resty.New()
	}

	client.SetUserAgent(DefaultUserAgent)

	baseURL, baseURLExists := os.LookupEnv("SITEHOST_URL")

	if baseURLExists {
		client.SetBaseURL(baseURL)
	}

	client.SetAPIVersion(APIVersion) // this call to set BaseURL

	client.
		SetRetryWaitTime((1000 * APISecondsPerPoll) * time.Millisecond).
		SetPollDelay(1000 * APISecondsPerPoll).
		SetRetries().
		SetDebug(envDebug)

	addResources(&client)

	return client
}

func addResources(client *Client) {
	resources := map[string]*Resource{
		serversName: NewResource(client, serversName, serversEndpoint, Server{}, ServersResponse{}),
	}

	client.resources = resources

	client.Servers = resources[serversName]
}
