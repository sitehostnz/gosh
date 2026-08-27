package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// rateLimitMarker identifies a throttled response.
//
// Matching on the message text is unpleasant and deliberate: the API
// signals a rate limit with **HTTP 500**, not 429, so the status code
// cannot be used to tell "slow down" from "something broke". Until that
// changes there is nothing else to match on.
//
// The fragility is real — an edit to the upstream wording silently
// disables this retry. TestRateLimit_MarkerIsPinned exists to make that
// break loudly in review rather than in production. Raising the status
// code to 429 has been requested; see the rate-limit ticket in the
// project notes.
const rateLimitMarker = "exceeded the number of requests per second"

// Retry defaults.
//
// The upstream limiter is a Redis counter over a rolling second, with a
// default allowance of 10 requests per second per reseller (not per
// key), configurable per reseller. So the useful backoff is short — the
// window it is waiting out is a second, not a minute — and a long
// exponential climb would only add latency.
const (
	defaultRateLimitAttempts = 4
	defaultRateLimitBackoff  = 250 * time.Millisecond
	maxRateLimitBackoff      = time.Second
)

// RateLimitError reports that a request was still being throttled after
// every attempt was used.
//
// It wraps the last response error, so callers that already inspect
// *models.ErrorResponse keep working through errors.As.
type RateLimitError struct {
	// Attempts is how many requests were made, including the first.
	Attempts int

	// Err is the last error returned by the API.
	Err error
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited after %d attempt(s): %v", e.Attempts, e.Err)
}

// Unwrap exposes the underlying API error.
func (e *RateLimitError) Unwrap() error { return e.Err }

// IsRateLimited reports whether err is, or wraps, a rate-limit
// rejection.
//
// It is true both for a [RateLimitError] — retries exhausted — and for a
// single throttled response when retrying is switched off, so callers
// can test one thing regardless of configuration.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	var rle *RateLimitError
	if errors.As(err, &rle) {
		return true
	}
	return isRateLimitMessage(err.Error())
}

// isRateLimitMessage reports whether a message is the throttle
// rejection.
func isRateLimitMessage(msg string) bool {
	return strings.Contains(strings.ToLower(msg), rateLimitMarker)
}

// SetRateLimitRetries configures how many times a request is attempted
// when the API reports a rate limit.
//
// The count includes the first attempt, so 1 disables retrying and 0 is
// treated the same way. The default is 4.
//
// Retrying is safe, including for requests that create things: the
// limit is applied after the key is authenticated but before the
// request is dispatched, so a throttled call never reaches the handler
// and cannot have had any effect.
//
// Internal resellers are configured with no limit at all and never see
// these responses; setting 1 for such a key costs nothing either way.
func SetRateLimitRetries(attempts int) ClientOpt {
	return func(c *Client) error {
		if attempts < 1 {
			attempts = 1
		}
		c.rateLimitAttempts = attempts
		return nil
	}
}

// SetRateLimitBackoff overrides the initial wait between throttled
// attempts. Each subsequent wait doubles, capped at one second.
func SetRateLimitBackoff(d time.Duration) ClientOpt {
	return func(c *Client) error {
		if d < 0 {
			return fmt.Errorf("api.SetRateLimitBackoff: backoff cannot be negative")
		}
		c.rateLimitBackoff = d
		return nil
	}
}

// rateLimitWait sleeps before the next attempt, honouring context
// cancellation so a caller that gives up is not held by a backoff.
func rateLimitWait(ctx context.Context, attempt int, base time.Duration) error {
	if base <= 0 {
		return nil
	}
	wait := base << (attempt - 1)
	if wait > maxRateLimitBackoff {
		wait = maxRateLimitBackoff
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
