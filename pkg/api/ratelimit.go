package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/models"
)

// rateLimitMarker identifies a throttled response.
//
// Matching on the message text is unpleasant and deliberate: the API
// signals a rate limit with HTTP 500, not 429, so the status code
// cannot be used to tell "slow down" from "something broke". Until that
// changes there is nothing else to match on.
//
// # The fragility, and what is and is not covered
//
// An edit to the API's wording silently disables this retry, and no
// unit test can detect that: any test here compares our constant
// against a copy of the message that also lives here, so both sides
// move together and CI stays green while the feature stops working.
//
// TestRateLimit_MarkerIsPinned guards something narrower and worth
// being precise about — that this constant is not edited without
// deliberation. It is a local-edit guard, not a drift guard. Detecting
// real drift needs a contract test against the live API, which does
// not exist yet.
//
// The exposure is reduced instead: rather than matching one long exact
// phrase, which a doubled space or a hyphen would break, the two
// stable words are matched independently after collapsing whitespace.
//
// Signalling this with 429 and Retry-After has been raised with the
// platform team. If that lands, all of this can be removed.
const (
	rateLimitMarkerA = "exceeded"
	rateLimitMarkerB = "requests per second"
)

// Retry defaults.
//
// The limiter is a counter over a rolling second, with a default
// allowance of 10 requests per second per reseller rather than per
// key, and it is configurable. So the useful backoff is short — the
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

// IsRateLimited reports whether an error is the API's rate-limit
// rejection.
//
// [Client.Do] returns a [RateLimitError] for a throttled request in
// every configuration, including when retrying is switched off — in
// that case with Attempts of 1. It does not return a bare
// *models.ErrorResponse for a throttle, so a caller matching on the
// concrete type with a type assertion or a type switch should move to
// errors.As. The underlying API error stays reachable that way.
//
// This function additionally recognises a throttled *models.ErrorResponse
// reached by some other route, so one test covers both.
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
//
// Both words are required, independently, after collapsing runs of
// whitespace. Matching the full phrase exactly would be broken by a
// doubled space or a hyphen upstream, which are cosmetic edits nobody
// would think to announce.
func isRateLimitMessage(msg string) bool {
	flat := strings.ToLower(strings.Join(strings.Fields(msg), " "))
	return strings.Contains(flat, rateLimitMarkerA) &&
		strings.Contains(flat, rateLimitMarkerB)
}

// isThrottled reports whether an error is the limiter's own rejection.
//
// This is what the retry loop decides on, and it is deliberately
// narrower than [IsRateLimited].
//
// The safety argument for retrying a request that creates something is
// that the limit is applied after the key is authenticated but before
// the request is dispatched, so a throttled call never reached the
// handler. That argument holds for the limiter's response and nothing
// else, so the guard has to be anchored to that response rather than to
// the text of whatever error came back.
//
// Three things follow, and each of them matters:
//
//   - It must be a parsed API error. Matching the rendered string of an
//     arbitrary error means matching the method and URL too, so a
//     validation failure echoing a caller's own input could trip it.
//     The SDK sends plenty of free text — record contents, stack names,
//     key comments.
//   - It must carry the status the limiter uses. Nothing else on 500
//     should be retried.
//   - The phrase must come from the message field, not from anywhere
//     else in the rendered error.
//
// Transport errors are excluded by all three, which is the important
// part. A timeout or a reset is the one case where the request may well
// have reached the handler and run, so it is exactly where the "never
// reached the handler" guarantee is false — and where retrying a create
// could make two.
func isThrottled(err error) bool {
	var apiErr *models.ErrorResponse
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Response != nil && apiErr.Response.StatusCode != http.StatusInternalServerError {
		return false
	}
	return isRateLimitMessage(apiErr.Message)
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
// Setting 1 disables retrying. That costs nothing for a caller whose
// allowance is high enough that it never trips the limit.
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
	wait := backoffFor(attempt, base)

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// backoffFor computes the wait before a given attempt.
//
// Separated from the sleeping so the schedule can be asserted without
// a test having to wait for it. Nothing pinned the progression before,
// which is why two faults in it went unnoticed: a configured value
// above the ceiling was silently discarded, and a large attempt count
// overflowed.
//
// Double up to the ceiling rather than shifting and clamping after.
//
// The shift form had two faults. It overflowed: the attempt count is
// caller-controlled and unbounded, so a large one wrapped the duration
// negative — the cap was then skipped, the timer fired immediately, and
// a caller asking to be more patient with a rate limiter got a
// zero-delay hot loop against it. It also panicked on a non-positive
// attempt, which was unreachable only because its single caller
// guarded it, leaving the safety outside the function that needed it.
//
// The ceiling applies to the growth, not to the caller's own value: a
// caller who deliberately chooses a gentler first wait gets it.
func backoffFor(attempt int, base time.Duration) time.Duration {
	wait := base
	for i := 1; i < attempt; i++ {
		if wait >= maxRateLimitBackoff {
			return maxRateLimitBackoff
		}
		wait *= 2
	}
	return wait
}
