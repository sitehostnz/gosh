package recorder

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
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
// It is not a redactor for arbitrary secrets. Values are safe because
// everything is discarded rather than because anything is recognised,
// so do not extend that with exceptions that let real values through.
//
// Object keys are the exception, and they are a judgement rather than a
// guarantee. A struct's keys are its shape and must survive; a map
// keyed by a customer's domain or address must not. Nothing in the JSON
// distinguishes the two, so [scrubKey] applies the leaf rules to keys
// that look like data and keeps the rest. It will be wrong in both
// directions on an endpoint nobody has recorded yet — read a fixture
// before committing it, particularly from an endpoint whose Return is
// a map.
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
		return scrubObject(t, depth)
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

// hostname matches a dotted name ending in a letters-only label, which
// is what a domain looks like and what a schema key does not.
var hostname = regexp.MustCompile(`^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*\.[a-zA-Z]{2,}$`)

// scrubKey replaces an object key that looks like data rather than
// schema.
//
// Keys were passed through untouched, which is right for a struct —
// there the key set is the shape, and replacing it would destroy the
// thing a fixture exists to record. It is wrong for the responses this
// API keys by customer data: redirect.ListRedirects returns
// map[domain]map[sourceURL]Rule, server.ListAllocatedIPs returns a map
// keyed by the address itself. Scrubbing such a response replaced the
// IPAddr value and left the address in the key above it.
//
// So the leaf rules are applied to keys that look like data —
// addresses, timestamps, digit strings, anything with an @, anything
// shaped like a hostname — and every other key is treated as schema
// and kept.
//
// This guesses, and it will guess wrong in both directions: an
// endpoint keyed by something none of these patterns match still
// leaks, and a schema key that happens to look like a hostname is
// destroyed. It errs towards discarding, which is the direction the
// package's own doc claims. Where the stakes are higher than a guess,
// check the output before committing it — which is what the "read the
// fixture before you commit it" line in Scrub's doc is for.
func scrubKey(k string) string {
	switch {
	case digitsOnly.MatchString(k),
		timestamp.MatchString(k),
		ipv4.MatchString(k),
		strings.Contains(k, "@"),
		hostname.MatchString(k):
		return scrubString(k)
	default:
		return k
	}
}

// scrubObject scrubs an object's values, and its keys where they look
// like data rather than schema.
func scrubObject(t map[string]any, depth int) map[string]any {
	out := make(map[string]any, len(t))

	// Sorted, so a data key's replacement is stable between runs and a
	// fixture does not churn.
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var n int
	for _, k := range keys {
		replaced := scrubKey(k)
		if replaced != k {
			// Number the replacements. Two data keys scrubbing to the
			// same placeholder would collide and silently collapse the
			// map, losing the entry count — and the count is part of
			// the shape a fixture exists to record.
			n++
			replaced = fmt.Sprintf("%s-%d", replaced, n)
		}
		out[replaced] = scrubValue(t[k], depth+1)
	}
	return out
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
