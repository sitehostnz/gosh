package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sitehostnz/gosh/pkg/models"
)

// rateLimitBody is the response the API sends when throttled. Note the
// 500: the status code cannot be used to identify it.
const rateLimitBody = `{"status":false,"msg":"You have exceeded the number of requests per second for this key. Please try again soon."}`

// TestRateLimit_MarkerIsPinned guards against an undeliberated local
// edit to the marker constants.
//
// It cannot detect an upstream rewording, and it is worth being exact
// about that: both sides of the comparison live in this repository, so
// they move together, CI stays green, and the retry silently stops
// working. Real drift detection needs a contract test against the live
// API, which does not exist.
//
// What it can do that the tolerance test cannot is assert the exact
// full phrase, so an edit narrowing the tokens fails here specifically.
func TestRateLimit_MarkerIsPinned(t *testing.T) {
	t.Parallel()

	const upstream = "Error: You have exceeded the number of requests per second for this key. Please try again soon."
	if !isRateLimitMessage(upstream) {
		t.Fatal("rateLimitMarkerA/B no longer match the upstream message; retrying is now disabled")
	}
	// The exact phrase, so narrowing either constant is caught here and
	// not only by the looser tolerance test.
	for _, want := range []string{rateLimitMarkerA, rateLimitMarkerB} {
		if !strings.Contains(strings.ToLower(upstream), want) {
			t.Errorf("marker %q is no longer part of the upstream message", want)
		}
	}
	if isRateLimitMessage("Error: The image 'x' could not be found.") {
		t.Error("unrelated errors must not be treated as rate limits")
	}
}

// TestRateLimit_RetriesThenSucceeds covers the ordinary case: a couple
// of throttled responses followed by a good one.
func TestRateLimit_RetriesThenSucceeds(t *testing.T) {
	t.Parallel()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, rateLimitBody)
			return
		}
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"ok":true}}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitBackoff(time.Millisecond))
	req, err := c.NewRequest("GET", "thing.json", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	var out struct {
		Return struct{ OK bool } `json:"return"`
	}
	if err := c.Do(context.Background(), req, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
	if !out.Return.OK {
		t.Error("response was not decoded after the retry")
	}
}

// TestRateLimit_ExhaustedReturnsTypedError checks the caller can
// distinguish "throttled throughout" from any other failure.
func TestRateLimit_ExhaustedReturnsTypedError(t *testing.T) {
	t.Parallel()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, rateLimitBody)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitRetries(3), SetRateLimitBackoff(time.Millisecond))
	req, err := c.NewRequest("GET", "thing.json", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	err = c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("Do: expected an error when every attempt is throttled")
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited(%v) = false, want true", err)
	}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("error is %T, want *RateLimitError", err)
	}
	if rle.Attempts != 3 {
		t.Errorf("Attempts = %d, want 3", rle.Attempts)
	}
	// The underlying API error must remain reachable, so callers that
	// already inspect it keep working.
	var apiErr *models.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Error("the wrapped *models.ErrorResponse is no longer reachable via errors.As")
	}
}

// TestRateLimit_DisabledMakesOneAttempt confirms retrying can be turned
// off, and that a single throttled response is still recognisable.
func TestRateLimit_DisabledMakesOneAttempt(t *testing.T) {
	t.Parallel()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, rateLimitBody)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitRetries(1))
	req, err := c.NewRequest("GET", "thing.json", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	err = c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("Do: expected an error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 when retrying is disabled", calls)
	}
	// Do returns *RateLimitError here too, with Attempts of 1;
	// IsRateLimited recognises it either way.
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited(%v) = false, want true", err)
	}
}

// TestRateLimit_RetriesWritesWithBody is the case that matters most: a
// POST carrying a form body must be replayed intact, not truncated.
//
// Retrying a write is only sound because the limit is applied before the
// request is dispatched, so the first attempt cannot have taken effect.
func TestRateLimit_RetriesWritesWithBody(t *testing.T) {
	t.Parallel()
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(raw))
		w.Header().Set("Content-Type", "application/json")
		if len(bodies) < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, rateLimitBody)
			return
		}
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitBackoff(time.Millisecond))
	req, err := c.NewRequest("POST", "thing.json", "label=web&location=AKLNCT")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if err := c.Do(context.Background(), req, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("attempts = %d, want 2", len(bodies))
	}
	if bodies[0] != bodies[1] {
		t.Errorf("retry sent a different body:\n first: %q\nsecond: %q", bodies[0], bodies[1])
	}
	if !strings.Contains(bodies[1], "location=AKLNCT") {
		t.Errorf("replayed body lost its content: %q", bodies[1])
	}
}

// TestRateLimit_NonRateLimitErrorsAreNotRetried makes sure an ordinary
// failure still fails immediately.
func TestRateLimit_NonRateLimitErrorsAreNotRetried(t *testing.T) {
	t.Parallel()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":false,"msg":"Error: The image 'x' could not be found."}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitBackoff(time.Millisecond))
	req, err := c.NewRequest("GET", "thing.json", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if err := c.Do(context.Background(), req, nil); err == nil {
		t.Fatal("Do: expected an error")
	} else if IsRateLimited(err) {
		t.Errorf("IsRateLimited(%v) = true, want false", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 — ordinary errors must not be retried", calls)
	}
}

// TestRateLimit_ContextCancellationStopsBackoff checks a caller that
// gives up is not held by a pending backoff.
func TestRateLimit_ContextCancellationStopsBackoff(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, rateLimitBody)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, SetRateLimitBackoff(2*time.Second))
	req, err := c.NewRequest("GET", "thing.json", "")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = c.Do(ctx, req, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Do error = %v, want context.DeadlineExceeded", err)
	}
	// Both facts must survive. Returning only the context error leaves
	// a rate limit indistinguishable from a hung request, which is the
	// misdiagnosis this retry exists to prevent.
	if !IsRateLimited(err) {
		t.Errorf("Do error = %v, want the throttle still reachable alongside the timeout", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Do waited %s after the context expired; backoff is not honouring cancellation", elapsed)
	}
}

// newTestClient builds a client pointed at a test server.
func newTestClient(t *testing.T, baseURL string, opts ...ClientOpt) *Client {
	t.Helper()
	all := append([]ClientOpt{SetBaseURL(baseURL)}, opts...)
	c, err := New("k", "1", all...)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return c
}

// TestRateLimit_BackoffSchedule pins the progression and the ceiling.
//
// Nothing asserted this before, which is how two faults in the same
// four lines went unnoticed: a configured value above the ceiling was
// silently reduced to it, and a large attempt count overflowed the
// shift and produced a zero wait — turning "be more patient with a
// rate limiter" into a hot loop against one.
func TestRateLimit_BackoffSchedule(t *testing.T) {
	t.Parallel()

	const base = 250 * time.Millisecond

	cases := []struct {
		name    string
		attempt int
		base    time.Duration
		want    time.Duration
	}{
		{"first attempt waits the configured value", 1, base, base},
		{"second doubles", 2, base, 500 * time.Millisecond},
		{"third doubles again", 3, base, time.Second},
		{"fourth is held at the ceiling", 4, base, time.Second},

		// The caller's own value is honoured rather than clamped. A
		// batch job that would rather wait five seconds than press a
		// shared allowance gets five seconds.
		// The row whose absence let a half-fix look complete. Asserting
		// only attempt 1 hid that later attempts were still clamped back
		// to the ceiling, so the schedule decreased: 5s, 1s, 1s.
		{"a base above the ceiling is the caller's choice", 1, 5 * time.Second, 5 * time.Second},
		{"and holds on the second attempt", 2, 5 * time.Second, 5 * time.Second},
		{"and the third", 3, 5 * time.Second, 5 * time.Second},
		{"and does not decrease at any depth", 10, 5 * time.Second, 5 * time.Second},

		// Overflow. The shift form wrapped negative here, skipped the
		// cap, and fired the timer immediately.
		{"a large attempt count stays at the ceiling", 38, base, time.Second},
		{"and stays there", 59, base, time.Second},
		{"and at any count", 200, base, time.Second},

		// The loop form is safe for inputs its caller currently
		// prevents; the shift form panicked on them.
		{"a zero attempt does not panic", 0, base, base},
		{"nor a negative one", -3, base, base},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := backoffFor(tc.attempt, tc.base); got != tc.want {
				t.Errorf("backoffFor(%d, %v) = %v, want %v", tc.attempt, tc.base, got, tc.want)
			}
		})
	}
}

// TestRateLimit_SetterGuards covers the configuration paths.
func TestRateLimit_SetterGuards(t *testing.T) {
	t.Parallel()

	t.Run("zero retries is treated as one", func(t *testing.T) {
		t.Parallel()
		c, err := New("k", "1", SetRateLimitRetries(0))
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.rateLimitAttempts != 1 {
			t.Errorf("attempts = %d, want 1", c.rateLimitAttempts)
		}
	})

	t.Run("a negative backoff is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := New("k", "1", SetRateLimitBackoff(-time.Second)); err == nil {
			t.Error("expected an error for a negative backoff")
		}
	})

	t.Run("a backoff above the ceiling is kept, not silently reduced", func(t *testing.T) {
		t.Parallel()
		c, err := New("k", "1", SetRateLimitBackoff(5*time.Second))
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.rateLimitBackoff != 5*time.Second {
			t.Errorf("backoff = %v, want 5s — the caller's value must survive", c.rateLimitBackoff)
		}
	})
}

// TestIsRateLimited_Nil checks the obvious input.
func TestIsRateLimited_Nil(t *testing.T) {
	t.Parallel()
	if IsRateLimited(nil) {
		t.Error("IsRateLimited(nil) = true")
	}
}

// TestRateLimit_TransportErrorsAreNotRetried is the safety property.
//
// The argument for retrying a request that creates something is that
// the limit is applied before dispatch, so a throttled call never
// reached the handler. A transport error is the one case where the
// request may well have reached the handler and run — so it is exactly
// where that guarantee is false, and exactly where a retry could make
// two servers.
//
// The earlier implementation matched the phrase against the rendered
// string of any error, transport errors included.
func TestRateLimit_TransportErrorsAreNotRetried(t *testing.T) {
	t.Parallel()

	transport := fmt.Errorf(
		"Get \"https://api.example/thing.json\": read tcp: exceeded the number of requests per second for this key")
	if isThrottled(transport) {
		t.Error("a transport error was treated as a throttle; the pre-dispatch guarantee does not hold for it")
	}

	// And an API error on some other status is not a throttle either.
	other := &models.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusBadRequest},
		Message:  "exceeded the number of requests per second for this key",
	}
	if isThrottled(other) {
		t.Error("a 400 carrying the phrase was treated as a throttle")
	}

	throttle := &models.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusInternalServerError},
		Message:  "You have exceeded the number of requests per second for this key.",
	}
	if !isThrottled(throttle) {
		t.Error("the limiter's own response was not recognised")
	}
}

// TestRateLimit_MarkerToleratesCosmeticRewording checks the matcher
// survives the edits nobody would announce.
func TestRateLimit_MarkerToleratesCosmeticRewording(t *testing.T) {
	t.Parallel()

	for _, msg := range []string{
		"You have exceeded the number of requests per second for this key.",
		"You have  exceeded  the number of requests per second for this key.",
		"You have exceeded the number of\nrequests per second for this key.",
		"YOU HAVE EXCEEDED THE NUMBER OF REQUESTS PER SECOND FOR THIS KEY.",
	} {
		if !isRateLimitMessage(msg) {
			t.Errorf("did not match: %q", msg)
		}
	}
	for _, msg := range []string{
		"Please specify a valid domain name.",
		"exceeded your storage quota",
		"requests per second is the unit we use",
	} {
		if isRateLimitMessage(msg) {
			t.Errorf("matched an unrelated message: %q", msg)
		}
	}
}

// TestRateLimit_NonReplayableBodyIsAttemptedOnce covers the path that
// prevents a truncated second copy.
//
// I had said this needed a test seam or a hand-built request. It needs
// neither: build the request the normal way and clear GetBody, which is
// the state net/http leaves a request in when its body is an opaque
// reader.
func TestRateLimit_NonReplayableBodyIsAttemptedOnce(t *testing.T) {
	t.Parallel()

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, rateLimitBody)
	}))
	defer srv.Close()

	c, err := New("k", "1", SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	req, err := c.NewRequest("POST", "thing.json", "label=web")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.GetBody = nil

	err = c.Do(context.Background(), req, nil)
	if calls != 1 {
		t.Errorf("handler hit %d times, want 1 — an unreplayable body must not be sent twice", calls)
	}

	// The type contract has no exceptions: every throttled path reports
	// a *RateLimitError, so a caller told to migrate to errors.As is
	// not left with one shape that slips through.
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("err = %T, want *RateLimitError", err)
	}
	if rle.Attempts != 1 {
		t.Errorf("Attempts = %d, want 1 — one request was actually sent", rle.Attempts)
	}
	var apiErr *models.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Error("the underlying API error is no longer reachable")
	}
}

// TestRateLimit_BodyReplayFailureIsSurfaced covers the other branch:
// a GetBody that errors is a real I/O failure, not a rate limit.
func TestRateLimit_BodyReplayFailureIsSurfaced(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, rateLimitBody)
	}))
	defer srv.Close()

	c, err := New("k", "1", SetBaseURL(srv.URL), SetRateLimitBackoff(time.Millisecond))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	req, err := c.NewRequest("POST", "thing.json", "label=web")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	boom := errors.New("boom")
	req.GetBody = func() (io.ReadCloser, error) { return nil, boom }

	err = c.Do(context.Background(), req, nil)

	// The actual cause must survive. Relabelling an I/O failure as a
	// rate limit sends the caller looking in the wrong place.
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want the replay failure to be reachable", err)
	}
	// And the throttle stays reachable too, which is what errors.Join
	// was for.
	var apiErr *models.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Error("the throttle that caused the retry is no longer reachable")
	}
}

// TestRateLimitError_Rendering pins the string a caller sees in a log.
func TestRateLimitError_Rendering(t *testing.T) {
	t.Parallel()

	e := &RateLimitError{Attempts: 4, Err: errors.New("500 You have exceeded the number of requests per second")}
	got := e.Error()
	if !strings.Contains(got, "4") {
		t.Errorf("Error() = %q, want the attempt count", got)
	}
	if !strings.Contains(got, "exceeded the number of requests") {
		t.Errorf("Error() = %q, want the wrapped API message", got)
	}
	if !errors.Is(e, e.Err) {
		t.Error("the wrapped error is not reachable through errors.Is")
	}
}

// TestRateLimit_BackoffIsNeverDecreasingAndNeverOverTheCeiling asserts
// the property rather than sampling it.
//
// The table above sampled two bases and both were structurally immune
// to the defect it was written to catch: 250ms doubles exactly onto the
// ceiling so the guard fires rather than overshooting, and 5s makes the
// ceiling equal to itself so the loop returns on its first iteration
// and never doubles. Twice on this one function the row that mattered
// was the row that was missing, which is an argument against picking
// samples at all.
//
// These two properties are what a backoff is. Anything satisfying them
// is acceptable; anything violating them is not, whatever the numbers.
func TestRateLimit_BackoffIsNeverDecreasingAndNeverOverTheCeiling(t *testing.T) {
	t.Parallel()

	bases := []time.Duration{
		time.Millisecond, 100 * time.Millisecond, 250 * time.Millisecond,
		300 * time.Millisecond, 600 * time.Millisecond, 750 * time.Millisecond,
		999 * time.Millisecond, time.Second, 1500 * time.Millisecond,
		5 * time.Second, time.Minute,
	}

	for _, base := range bases {
		ceiling := maxRateLimitBackoff
		if base > ceiling {
			ceiling = base
		}

		var prev time.Duration
		for attempt := 1; attempt <= 10; attempt++ {
			got := backoffFor(attempt, base)

			if got > ceiling {
				t.Errorf("backoffFor(%d, %v) = %v, over the ceiling %v", attempt, base, got, ceiling)
			}
			if attempt > 1 && got < prev {
				t.Errorf("backoffFor(%d, %v) = %v, less than the previous wait %v — a schedule that decreases is not a backoff, and it shrinks at the moment the limiter has just proven it is still rejecting",
					attempt, base, got, prev)
			}
			if got <= 0 {
				t.Errorf("backoffFor(%d, %v) = %v, want a positive wait", attempt, base, got)
			}
			prev = got
		}

		// The caller's own value is the first wait, always.
		if got := backoffFor(1, base); got != base {
			t.Errorf("backoffFor(1, %v) = %v, want the configured value", base, got)
		}
	}
}
