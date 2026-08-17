package models

import (
	"fmt"
	"net/http"
	"net/url"
)

type (
	// APIResponse represents mutual API response values.
	APIResponse struct {
		Msg    string `json:"msg"`
		Status bool   `json:"status"`
	}

	// ErrorResponse reports the error caused by an API request.
	ErrorResponse struct {
		Response *http.Response `json:"-"`
		Message  string         `json:"msg"`
		Status   bool           `json:"status"`
	}
)

// Error returns a ErrorResponse message.
//
// The URL is redacted before it is formatted. The API key travels as a
// query parameter, so an unredacted URL puts the caller's credential
// into every error string — and error strings get wrapped, logged,
// shipped to log aggregators and pasted into tickets.
func (r *ErrorResponse) Error() string {
	if r.Response == nil || r.Response.Request == nil {
		return r.Message
	}
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, RedactURL(r.Response.Request.URL),
		r.Response.StatusCode, r.Message)
}

// credentialParams are query parameters whose values must never appear
// in an error, a log line, or anything else a human might forward.
//
// Deliberately just the API key. It is long-lived, grants the whole API,
// rides on every single request, and echoing it back tells the caller
// nothing they did not already supply — so redacting it costs no
// debugging information at all.
//
// Other secret-sounding parameters are NOT redacted, and that is a
// decision rather than an oversight. `token` on vnc/valid_token is the
// subject of the request: it is short-lived, scoped to one console, and
// it is the field you need to see when the call fails. An error reading
// `token=REDACTED is invalid` hides the input from the person who
// supplied it, which is not protection — it is a worse error.
//
// The rule this follows: redact what rides along on every request, not
// what the request is about.
var credentialParams = []string{"apikey", "api_key"}

// RedactURL returns u as a string with credential-bearing query
// parameters replaced by REDACTED. The original URL is not modified.
//
// Exported so callers building their own error text get the same
// treatment rather than reimplementing it — the alternative is that
// each caller redacts a slightly different set, which is how one gets
// missed.
func RedactURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	q := u.Query()
	redacted := false
	for _, k := range credentialParams {
		if q.Has(k) {
			q.Set(k, "REDACTED")
			redacted = true
		}
	}
	if !redacted {
		return u.String()
	}
	// Copy so a caller inspecting the request afterwards still sees the
	// real URL — redaction is for display, not for the request itself.
	c := *u
	c.RawQuery = q.Encode()
	return c.String()
}
