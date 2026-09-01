package shtypes

import (
	"strconv"
	"strings"
)

// MaybeBool is a simple wrapper unmarshalling json responses that could be either an int or an int in a string.
type MaybeBool bool

// UnmarshalJSON is a helper interface for dealing with things that may or may not be a string representing a bool, or  number representing a bool.
func (fi *MaybeBool) UnmarshalJSON(b []byte) error {
	v := strings.Trim(string(b), "\"")

	// null and empty decode as false rather than failing.
	//
	// strconv.ParseBool rejects both, and the error propagates out of
	// the whole response — so one null in a field a caller never reads
	// would break the endpoint for everyone. That is worse than the
	// behaviour without this type at all, where an unknown field is
	// simply dropped.
	//
	// Whether the API means false by null is not established; what is
	// established is that failing the decode is the wrong answer
	// either way.
	if v == "null" || v == "" {
		*fi = false
		return nil
	}

	maybeBool, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}

	*fi = MaybeBool(maybeBool)
	return nil
}
