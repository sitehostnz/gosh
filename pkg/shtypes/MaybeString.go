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
	if len(b) == 0 || string(b) == "null" {
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
	return &json.UnmarshalTypeError{Value: string(b), Type: reflect.TypeOf(MaybeString(""))}
}

// String returns the underlying string value.
func (fi MaybeString) String() string { return string(fi) }

// Int parses the value as an int.
func (fi MaybeString) Int() (int, error) { return strconv.Atoi(string(fi)) }
