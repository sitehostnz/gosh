package shtypes

import (
	"strconv"
	"strings"
)

// MaybeBool is a simple wrapper unmarshalling json responses that could be either an int or an int in a string.
type MaybeBool bool

// UnmarshalJSON is a helper interface for dealing with things that may or may not be a string representing a bool, or  number representing a bool.
func (fi *MaybeBool) UnmarshalJSON(b []byte) error {
	maybeBool, err := strconv.ParseBool(strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}

	*fi = MaybeBool(maybeBool)
	return nil
}
