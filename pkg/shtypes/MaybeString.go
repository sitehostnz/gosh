package shtypes

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
)

// MaybeString is for things that may or may not be strings, but should be
// represented as a string. The API surfaces the same id-like columns as
// strings in list views and as numbers in detail views (e.g. a list
// returns `"client_id": "1"` where the detail view returns
// `"client_id": 1`), so either form decodes into a Go string.
type MaybeString string

// UnmarshalJSON accepts a JSON string, a JSON number, or `null`.
// Anything else (bool, array, object) returns an error rather than a
// silently stringified value. String escape sequences are decoded by the
// stdlib, so the stored value is the unescaped Go string.
func (fi *MaybeString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	// "null", absent — and "[]", which is PHP's json_encode emitting an
	// empty associative array for "no value". A wire that says "nothing"
	// must decode as empty rather than fail the whole response; several
	// downstream optional fields arrive exactly this way. "{}" is
	// deliberately NOT here, matching IsEmptyMapShape's definition: this
	// serialiser does not produce it for an empty value, so meeting one
	// means the shape genuinely is not scalar and should error.
	if len(b) == 0 || string(b) == "null" || string(b) == "[]" {
		*fi = ""
		return nil
	}
	// A JSON string first: encoding/json handles escapes and Unicode.
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*fi = MaybeString(str)
		return nil
	}
	// Fall back to a JSON number, preserving its textual form.
	var num json.Number
	if err := json.Unmarshal(b, &num); err == nil {
		*fi = MaybeString(num.String())
		return nil
	}
	// Describe the payload's TYPE, never its content — error strings get
	// logged and forwarded, and echoing rejected bytes would put wire
	// content into them (the same rule as the API-key redaction).
	kind := "invalid JSON"
	switch b[0] {
	case 't', 'f':
		kind = "bool"
	case '[':
		kind = "array"
	case '{':
		kind = "object"
	}
	return &json.UnmarshalTypeError{Value: kind, Type: reflect.TypeOf(MaybeString(""))}
}

// String returns the underlying string value.
func (fi MaybeString) String() string { return string(fi) }

// Int parses the value as an int.
func (fi MaybeString) Int() (int, error) { return strconv.Atoi(string(fi)) }
