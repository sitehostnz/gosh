package shtypes

// jsonNull is the JSON null literal, checked in several places here
// because this API uses null, "[]" and absence interchangeably for an
// empty value.
const jsonNull = "null"

// BoolToInt simple formatter for mapping a bool to a 1 or a zero.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IsEmptyMapShape reports whether a raw JSON payload is one of the
// shapes PHP produces for "no rows": absent, null, or an empty array.
//
// PHP's json_encode emits `[]` for an empty associative array rather
// than `{}`, because it cannot tell the difference between an empty map
// and an empty list. So an endpoint that normally returns a keyed object
// returns `[]` when there is nothing to return, and unmarshalling into a
// map or struct fails on a response that is entirely valid.
//
// Callers use it to short-circuit before decoding:
//
//	if shtypes.IsEmptyMapShape(raw) {
//	    return nil // no rows, not an error
//	}
//
// Lives here rather than in a single endpoint package because it is a
// property of the serialiser, not of any one response — it can appear
// wherever a map-shaped return can be empty.
func IsEmptyMapShape(raw []byte) bool {
	s := string(raw)
	return len(raw) == 0 || s == jsonNull || s == "[]"
}
