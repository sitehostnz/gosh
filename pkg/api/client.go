// Package api provides the functions to work with SiteHost API.
//
// # Rate limiting
//
// The API applies a per-second request limit. The allowance is not
// published and varies, so a client must not assume a particular
// figure.
//
// Exceeding it returns HTTP 500 with:
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
// Retrying is sound even for requests that create things, because the
// limit is enforced before the request reaches the handler — so a
// throttled call cannot have had an effect.
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
// The API signals a rate limit with HTTP 500 and a message rather than
// with 429, which makes it indistinguishable from a server error by
// status code alone. Left
// alone, a client reads that as "the operation failed" — and in a
// provisioning flow that means concluding a build failed when it never
// started.
//
// So a throttled response is retried with a short backoff. This is safe
// even for requests that create things: a throttled request is rejected
// without being processed, so the first attempt cannot have had any
// effect.
//
// Retrying needs the request body to be replayable. Bodies built by
// NewRequest are, because net/http populates GetBody for the reader
// types it uses; a request carrying a body it cannot replay is
// attempted once and reported as a [RateLimitError] with an attempt
// count of 1, rather than silently sending a truncated second copy.
//
// When every attempt is throttled the result is a [RateLimitError]
// wrapping the last API error. Use [IsRateLimited] to test for it.
// Configure with [SetRateLimitRetries] and [SetRateLimitBackoff].
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	attempts := c.rateLimitAttempts
	if attempts < 1 {
		attempts = 1
	}

	// Apply the caller's context to the request, not only to the
	// backoff. Wiring it into the wait alone left the parameter
	// half-working: cancelling aborted a pending sleep and did nothing
	// to an in-flight request, while the doc and a test both said
	// cancellation was honoured. Visibly ignored would have been more
	// honest than that.
	// A nil context used to be harmless here, because the parameter was
	// discarded. WithContext panics on nil, so accept it as Background
	// rather than turning a previously working call into a panic in a
	// published SDK.
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.WithContext(ctx)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			replayable, err := rewindBody(req)
			switch {
			case err != nil:
				// A real I/O failure re-reading the body. Reporting
				// this as a rate limit sends the caller looking in
				// entirely the wrong place, so both are surfaced.
				return errors.Join(lastErr, err)
			case !replayable:
				// A body that cannot be replayed is attempted once
				// rather than risking a truncated second copy. Still
				// reported as a RateLimitError, so the documented type
				// contract holds on every throttled path — with the
				// attempt count actually made, which is one fewer than
				// the attempt about to be abandoned.
				return &RateLimitError{Attempts: attempt - 1, Err: lastErr}
			}
			if err := rateLimitWait(ctx, attempt-1, c.rateLimitBackoff); err != nil {
				// Both facts matter: the caller ran out of time, and
				// the reason it was waiting was a throttle. Returning
				// only the context error leaves a rate limit
				// indistinguishable from a hung request, which is the
				// misdiagnosis this retry exists to prevent.
				return errors.Join(lastErr, err)
			}
		}

		err := c.do(req, v)
		if err == nil {
			return nil
		}
		if !isThrottled(err) {
			return err
		}
		lastErr = err
	}

	return &RateLimitError{Attempts: attempts, Err: lastErr}
}

// rewindBody restores a request body for another attempt.
//
// It returns whether the request can be attempted again, and separately
// any error from re-reading the body. The two are different situations
// and were previously collapsed into one: the caller discarded the
// error and returned the earlier throttle error, so a genuine I/O
// failure was reported as a rate limit. The message that was supposed
// to explain it was formatted and then unconditionally thrown away,
// which also made its attempt parameter dead.
func rewindBody(req *http.Request) (replayable bool, err error) {
	if req.Body == nil && req.GetBody == nil {
		return true, nil // nothing to replay
	}
	if req.GetBody == nil {
		return false, nil
	}
	body, err := req.GetBody()
	if err != nil {
		return false, err
	}
	req.Body = body
	return true, nil
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
