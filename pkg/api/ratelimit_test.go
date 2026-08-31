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

// TestRateLimit_MarkerIsPinned guards the one fragile part of this
// feature.
//
// The API signals a rate limit with HTTP 500 and a message, so the only
// thing to match on is the wording. If upstream edits it, the retry
// silently stops working and throttled calls start surfacing as failed
// operations again. This test fails loudly instead.
func TestRateLimit_MarkerIsPinned(t *testing.T) {
	t.Parallel()
	if !isRateLimitMessage("Error: You have exceeded the number of requests per second for this key. Please try again soon.") {
		t.Fatal("the upstream rate-limit message no longer matches rateLimitMarker; retrying is now disabled")
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
	if err := c.Do(ctx, req, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Do error = %v, want context.DeadlineExceeded", err)
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
		{"a base above the ceiling is the caller's choice", 1, 5 * time.Second, 5 * time.Second},

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
