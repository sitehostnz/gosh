package recorder

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// maxArrayElements caps how many elements of an array survive scrubbing.
//
// A fixture needs enough elements to prove the decoder handles a list,
// and enough to show whether elements vary in shape. Beyond that, extra
// elements are weight: an eighty-image catalogue makes a fixture nobody
// reads and a diff nobody reviews.
const maxArrayElements = 2

var (
	digitsOnly = regexp.MustCompile(`^\d+$`)
	timestamp  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}([ T]\d{2}:\d{2}:\d{2})?`)
	ipv4       = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
	ipv6       = regexp.MustCompile(`^[0-9a-fA-F:]{6,}$`)
)

// Scrub rewrites a recorded response body so it can be committed as a
// fixture, replacing every value while preserving the structure.
//
// # What it keeps, and why that is the whole point
//
// It keeps which keys are present, how they nest, and the JSON type of
// every leaf. It throws away the values.
//
// That split is deliberate, because the values were never what a
// fixture was for. The bugs a fixture catches are shape bugs: a field
// the API sends as a map that we declared a bool, a number that arrives
// quoted as a string, a key that is absent rather than null. All of
// those survive scrubbing intact. Meanwhile the things that must not
// reach a public repository — customer database names, usernames, home
// directories, key material, addresses — are exactly the values.
//
// A string is replaced with a placeholder of the same shape, so that a
// field the API sends as a quoted integer stays a quoted integer and a
// timestamp stays parseable. Numbers keep their type but not their
// magnitude. Booleans and nulls pass through, since neither carries
// anything and both are load-bearing: null-versus-absent is a
// distinction a hand-written fixture routinely gets wrong.
//
// # What it does not do
//
// It is not a redactor for arbitrary secrets. It is safe because it
// discards everything rather than because it recognises anything, so do
// not extend it with exceptions that let real values through.
func Scrub(raw []byte) ([]byte, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber() // keep 1 and 1.0 distinguishable

	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("scrub: %w", err)
	}

	out, err := json.MarshalIndent(scrubValue(doc, 0), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("scrub: %w", err)
	}
	return append(out, '\n'), nil
}

// scrubValue walks one JSON value. counter varies the placeholders so a
// fixture does not collapse into a page of identical strings.
func scrubValue(v any, depth int) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = scrubValue(val, depth+1)
		}
		return out
	case []any:
		n := len(t)
		if n > maxArrayElements {
			n = maxArrayElements
		}
		out := make([]any, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, scrubValue(t[i], depth+1))
		}
		return out
	case string:
		return scrubString(t)
	case json.Number:
		// Preserve integer-versus-float, which decoders care about.
		if strings.ContainsAny(t.String(), ".eE") {
			return json.Number("1.5")
		}
		return json.Number("1")
	default:
		// bool and nil, both of which say something and reveal nothing.
		return v
	}
}

// scrubString replaces a string with a placeholder of the same shape.
func scrubString(s string) string {
	switch {
	case s == "":
		// An empty string is a fact about the API, not a value.
		return ""
	case digitsOnly.MatchString(s):
		return "1"
	case timestamp.MatchString(s):
		return "2026-01-01 00:00:00"
	case ipv4.MatchString(s):
		// TEST-NET-3, reserved for documentation by RFC 5737.
		return "203.0.113.1"
	case ipv6.MatchString(s) && strings.Contains(s, ":"):
		// Reserved for documentation by RFC 3849.
		return "2001:db8::1"
	case strings.HasPrefix(s, "ssh-") || strings.HasPrefix(s, "ecdsa-"):
		return "ssh-ed25519 AAAA scrubbed"
	default:
		return "scrubbed"
	}
}
