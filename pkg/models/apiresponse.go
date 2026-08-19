package models

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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
	if r.Response == nil {
		return r.Message
	}
	if r.Response.Request == nil {
		// No request to describe, but the status code is real information
		// already in hand — dropping it would make the error worse than
		// it needs to be.
		return fmt.Sprintf("%d %v", r.Response.StatusCode, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, RedactURL(r.Response.Request.URL),
		r.Response.StatusCode, r.Message)
}

// credentialRe matches the API-key query parameter for redaction.
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
//
// Case-insensitive, because the point of exporting RedactURL is that
// callers building their own requests get the same treatment — and a
// caller who spells it APIKey= must not sail through the redaction.
var credentialRe = regexp.MustCompile(`(?i)((?:^|[&?])(?:apikey|api_key)=)[^&]*`)

// RedactURL returns u as a string with the API-key query parameter
// replaced by REDACTED. The original URL is not modified.
//
// It rewrites RawQuery textually rather than round-tripping through
// url.Values, so every other parameter survives byte-for-byte: original
// order, duplicates, and valueless keys all preserved. A redacted error
// should differ from the real URL by exactly one value — this repo added
// pkg/net.Encode precisely because Values.Encode re-sorts, and the error
// path should not reintroduce that.
//
// Exported so callers building their own error text get the same
// treatment rather than reimplementing it — the alternative is that
// each caller redacts a slightly different set, which is how one gets
// missed.
func RedactURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	if !credentialRe.MatchString(u.RawQuery) {
		return u.String()
	}
	// Copy so a caller inspecting the request afterwards still sees the
	// real URL — redaction is for display, not for the request itself.
	c := *u
	c.RawQuery = credentialRe.ReplaceAllString(u.RawQuery, "${1}REDACTED")
	return c.String()
}
