package shtypes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// MaybeBoolMap is a per-key flag map that tolerates a bare scalar in
// place of the object.
//
// This API answers some per-component fields with an object keyed by
// the component it is reporting on:
//
//	{"return": {"disk": {"scsi0": true}}}
//
// and there is reason to think it can answer the same field with a
// plain bool when there is nothing to key on — no component of that
// kind was requested, or the whole request was refused. Its sibling
// fields in the same object are plain bools, which is where the
// original (wrong) bool declaration came from.
//
// Declaring the field as a map alone traded one decode failure for
// another pointing the other way: "cannot unmarshal bool into Go
// struct field ... of type map[string]bool", which would replace the
// API's own message and leave a caller debugging a JSON type error
// instead of reading "Please specify a valid disk label."
//
// So both shapes decode:
//
//   - an object decodes to the map it describes;
//   - true decodes to an empty non-nil map, meaning accepted with
//     nothing to enumerate;
//   - false and null decode to a nil map, meaning not accepted.
//
// Use [MaybeBoolMap.Accepted] rather than testing the map directly, so
// that the distinction between the two empty cases stays in one place.
type MaybeBoolMap map[string]bool

// UnmarshalJSON accepts either the object form or a scalar bool.
func (m *MaybeBoolMap) UnmarshalJSON(b []byte) error {
	trimmed := strings.TrimSpace(string(b))
	if trimmed == jsonNull {
		*m = nil
		return nil
	}

	if strings.HasPrefix(trimmed, "{") {
		var raw map[string]bool
		if err := json.Unmarshal(b, &raw); err != nil {
			return fmt.Errorf("shtypes: MaybeBoolMap: %w", err)
		}
		*m = raw
		return nil
	}

	// A quoted or bare bool. Anything else is a genuine surprise and
	// should surface rather than be swallowed.
	flag, err := strconv.ParseBool(strings.Trim(trimmed, `"`))
	if err != nil {
		return fmt.Errorf("shtypes: MaybeBoolMap: want an object or a bool, got %s", trimmed)
	}
	if flag {
		*m = MaybeBoolMap{}
		return nil
	}
	*m = nil
	return nil
}

// Accepted reports whether the component was accepted at all, treating
// the scalar and object forms alike. A non-nil map is an acceptance,
// even when it enumerates nothing.
func (m MaybeBoolMap) Accepted() bool { return m != nil }

// AcceptedKey reports whether a named component was accepted.
//
// It answers false for a scalar acceptance, which enumerates nothing —
// so check [MaybeBoolMap.Accepted] first when the distinction between
// "not accepted" and "accepted without detail" matters.
func (m MaybeBoolMap) AcceptedKey(key string) bool { return m[key] }
