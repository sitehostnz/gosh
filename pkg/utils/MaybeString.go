package utils

import "strings"

type MaybeString string

// UnmarshalJSON is a helper interface for dealing with numbers that should be strings, but sometimes come through as number
func (fi *MaybeString) UnmarshalJSON(b []byte) error {
	*fi = MaybeString(strings.Trim(string(string(b)), "\""))
	return nil
}
