package types

import "strings"

// MaybeString is for things that may or may not be strings, but should be represented as a string.
type MaybeString string

// UnmarshalJSON is a helper interface for dealing with numbers that should be strings, but sometimes come through as number.
func (fi *MaybeString) UnmarshalJSON(b []byte) error {
	*fi = MaybeString(strings.Trim(string(b), "\""))
	return nil
}
