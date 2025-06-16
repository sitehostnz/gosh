package shtypes

// BoolToInt simple formatter for mapping a bool to a 1 or a zero.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
