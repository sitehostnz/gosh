// Package api provides the functions to work with SiteHost API.
//
// # Rate limiting
//
// The API limits requests per **reseller** — the owner of the API key,
// not the individual key. The default allowance is 10 requests per
// second, held as a counter over a rolling second, and it is
// configurable per reseller: internal resellers are unlimited, and a
// reseller can be set to zero to block them entirely. So a client
// should not hardcode the default.
//
// Exceeding the limit returns **HTTP 500** with:
//
//	You have exceeded the number of requests per second for this key.
//	Please try again soon.
//
// The status code is the problem. A rate limit is indistinguishable
// from a server error by status alone, so a client that treats 500 as
// "the operation failed" draws the wrong conclusion — and in a
// provisioning flow that means reporting a failed build that never
// started, or retrying a create and making two.
//
// [Client.Do] therefore retries throttled requests with a short
// backoff, and reports [RateLimitError] if every attempt is throttled.
// Retrying is sound even for requests that create things: the limit is
// applied after the key is authenticated but before the request is
// dispatched, so a throttled call never reaches the handler.
//
// Tune it with [SetRateLimitRetries] and [SetRateLimitBackoff], and test
// for it with [IsRateLimited].
//
// Polling a job in a tight loop is the usual way to trip the limit.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

const (
	defaultBaseURL            = "https://api.sitehost.nz"
	defaultVersion            = "1.5"
	userAgent                 = "gosh"
	defaultMediaType          = "application/x-www-form-urlencoded"
	defaultCustomImageGitHost = "gitlab-clients.sitehost.co.nz"
)

type (
	// Client is a wrapper around the http client to manages communication with SiteHost API.
	Client struct {
		client *http.Client
		models.ClientBase

		// rateLimitAttempts and rateLimitBackoff control retrying of
		// throttled requests. See SetRateLimitRetries.
		rateLimitAttempts int
		rateLimitBackoff  time.Duration
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

	// doing this, so we can kinda hope to preserve the order...
	// really client id and what not should perhaps come higher up the food chain,
	// and only default if not in the request.

	keys := []string{"apikey", "client_id"}
	values := make(url.Values)
	values.Add("apikey", c.APIKey)
	values.Add("client_id", c.ClientID)

	q := strings.Join(
		[]string{net.Encode(values, keys), u.RawQuery},
		"&",
	)

	u.RawQuery = q

	req, err := http.NewRequestWithContext(context.Background(), method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Every body this SDK sends is form-encoded, not JSON. The line
	// setting application/json here was immediately overwritten by the
	// next one, so it never had an effect — but it read as though the
	// SDK sometimes sends JSON, which it does not.
	if body != "" {
		req.Header.Set("Content-Type", defaultMediaType)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// Do sends an API Request and returns the response.
//
// The API response is checked to see if it was a successful call.
// A successful call is then checked to see if we have a Status true.
//
// # Throttled requests are retried
//
// The API rate-limits per reseller and signals it with HTTP 500 and a
// "you have exceeded the number of requests per second" message, which
// is indistinguishable from a server error by status code alone. Left
// alone, a client reads that as "the operation failed" — and in a
// provisioning flow that means concluding a build failed when it never
// started.
//
// So a throttled response is retried with a short backoff. This is safe
// even for requests that create things: the limit is applied after the
// key is authenticated but before the request is dispatched, so a
// throttled call never reached the handler.
//
// Retrying needs the request body to be replayable. Bodies built by
// NewRequest are, because net/http populates GetBody for the reader
// types it uses; a request carrying a body it cannot replay is attempted
// once and returned as-is rather than silently sending a truncated
// second copy.
//
// When every attempt is throttled the result is a [RateLimitError]
// wrapping the last API error. Use [IsRateLimited] to test for it.
// Configure with [SetRateLimitRetries] and [SetRateLimitBackoff].
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	attempts := c.rateLimitAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			if err := c.rewindBody(req, attempt); err != nil {
				return lastErr
			}
			if err := rateLimitWait(ctx, attempt-1, c.rateLimitBackoff); err != nil {
				return err
			}
		}

		err := c.do(req, v)
		if err == nil {
			return nil
		}
		if !isRateLimitMessage(err.Error()) {
			return err
		}
		lastErr = err
	}

	return &RateLimitError{Attempts: attempts, Err: lastErr}
}

// rewindBody restores a request body for another attempt, reporting an
// error when it cannot be replayed.
func (c *Client) rewindBody(req *http.Request, attempt int) error {
	if req.Body == nil && req.GetBody == nil {
		return nil // nothing to replay
	}
	if req.GetBody == nil {
		return fmt.Errorf("api: cannot retry attempt %d, request body is not replayable", attempt)
	}
	body, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = body
	return nil
}

// do performs a single request.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		// Transport errors (*url.Error: timeouts, DNS, TLS, resets) embed
		// the full request URL, query string included — and the API key
		// travels as a query parameter. net/http strips userinfo
		// passwords, never query parameters, so without this every
		// transport failure logs the caller's credential. These are the
		// errors most likely to reach a log aggregator.
		var uerr *url.Error
		if errors.As(err, &uerr) {
			uerr.URL = models.RedactURL(req.URL)
		}
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

// CheckResponse checks the API response for errors and returns them if present.
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
		client:            &http.Client{},
		rateLimitAttempts: defaultRateLimitAttempts,
		rateLimitBackoff:  defaultRateLimitBackoff,
		ClientBase: models.ClientBase{
			BaseURL:            baseURL,
			APIKey:             apiKey,
			ClientID:           clientID,
			UserAgent:          userAgent,
			CustomImageGitHost: defaultCustomImageGitHost,
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

// SetCustomImageGitHost overrides the GitLab host used by custom
// image helpers when constructing repository clone URLs. Default
// is "gitlab-clients.sitehost.co.nz".
func SetCustomImageGitHost(host string) ClientOpt {
	return func(c *Client) error {
		c.CustomImageGitHost = host
		return nil
	}
}
