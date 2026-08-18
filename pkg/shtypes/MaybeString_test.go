package shtypes

import "testing"

// TestMaybeStringUnmarshal covers both wire forms, the null/empty cases,
// escape handling, and the shapes that must be REJECTED rather than
// silently stringified. The rejection cases are the point of the
// json-delegating implementation: the previous strings.Trim version
// accepted `true` or `[1]` and stored them as text, which turns a wire
// contract violation into silently wrong data.
func TestMaybeStringUnmarshal(t *testing.T) {
	t.Parallel()
	ok := []struct {
		name, raw, want string
	}{
		{"string", `"1"`, "1"},
		{"number", `1`, "1"},
		{"big number textual form", `10633248768`, "10633248768"},
		{"float keeps textual form", `1.50`, "1.50"},
		{"null", `null`, ""},
		{"empty input", ``, ""},
		{"escaped quote", `"a\"b"`, `a"b`},
		{"unicode escape", `"é"`, "é"},
	}
	for _, tc := range ok {
		var v MaybeString
		if err := v.UnmarshalJSON([]byte(tc.raw)); err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if string(v) != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, v, tc.want)
		}
	}
	for _, raw := range []string{`true`, `false`, `[1]`, `{"a":1}`} {
		var v MaybeString
		if err := v.UnmarshalJSON([]byte(raw)); err == nil {
			t.Errorf("%q: want an error, got %q — a non-scalar silently stringified", raw, v)
		}
	}
}

func TestMaybeStringAccessors(t *testing.T) {
	t.Parallel()
	v := MaybeString("42")
	if v.String() != "42" {
		t.Errorf("String() = %q", v.String())
	}
	n, err := v.Int()
	if err != nil || n != 42 {
		t.Errorf("Int() = %d, %v", n, err)
	}
	if _, err := MaybeString("abc").Int(); err == nil {
		t.Error("Int() on non-numeric should error")
	}
}
