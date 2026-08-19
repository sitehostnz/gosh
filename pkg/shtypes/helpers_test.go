package shtypes

import "testing"

// TestIsEmptyMapShape covers the three forms PHP produces for "nothing
// to return", and confirms real payloads are not mistaken for them.
func TestIsEmptyMapShape(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{"absent", "", true},
		{"null", "null", true},
		{"php empty assoc array", "[]", true},
		{"populated map", `{"1":{"id":"1"}}`, false},
		{"empty map", "{}", false},
		{"populated array", `[{"id":"1"}]`, false},
		{"zero", "0", false},
		{"empty string", `""`, false},
	} {
		if got := IsEmptyMapShape([]byte(tc.raw)); got != tc.want {
			t.Errorf("%s: IsEmptyMapShape(%q) = %v, want %v", tc.name, tc.raw, got, tc.want)
		}
	}
}
