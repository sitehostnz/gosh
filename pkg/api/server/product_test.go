package server

import (
	"encoding/json"
	"testing"
)

// TestProductAttributes_MalformedIsReturnedNotSwallowed covers the
// error paths in ProductAttributes.UnmarshalJSON.
//
// The empty-list tolerance exists because PHP serialises an empty map
// as []. That tolerance has to stop at empty: a populated list, or a
// scalar, in the attributes position is a shape we do not understand,
// and understanding it wrongly is how a caller ends up with silently
// zeroed cores and RAM.
func TestProductAttributes_MalformedIsReturnedNotSwallowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"the empty-list shape PHP emits for an empty map", `[]`, false},
		{"the ordinary object", `{"cores":2,"ram":4}`, false},
		{"a populated list is not an empty map", `["cores"]`, true},
		{"a scalar is not an attribute set", `7`, true},
		{"a string is not an attribute set", `"cores"`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got ProductAttributes
			err := json.Unmarshal([]byte(tc.in), &got)
			if tc.wantErr && err == nil {
				t.Fatalf("Unmarshal(%s) = %#v, want an error rather than a silently empty value", tc.in, got)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Unmarshal(%s): %v", tc.in, err)
			}
		})
	}
}

// TestProductAttributes_ExtraDoesNotDuplicateTypedFields guards a
// subtlety: encoding/json matches field names case-insensitively, so a
// response using "Cores" would decode into the typed field, while Extra
// deletes only the lower-case key — leaving the same value in both
// places. Unlikely from this API, and cheap to pin.
func TestProductAttributes_ExtraDoesNotDuplicateTypedFields(t *testing.T) {
	t.Parallel()

	var got ProductAttributes
	if err := json.Unmarshal([]byte(`{"Cores":2,"colour":"blue"}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Cores != 2 {
		t.Errorf("Cores = %v, want 2 — encoding/json matches case-insensitively", got.Cores)
	}
	if _, duplicated := got.Extra["Cores"]; duplicated {
		t.Error(`Extra retains "Cores" as well as the typed field; a caller reading both sees it twice`)
	}
	if _, kept := got.Extra["colour"]; !kept {
		t.Error("Extra dropped an unknown key; keeping them is what Extra is for")
	}
}
