package shtypes_test

import (
	"encoding/json"
	"testing"

	"github.com/sitehostnz/gosh/pkg/shtypes"
)

func TestMaybeBoolMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       string
		accepted bool
		scsi0    bool
		wantErr  bool
	}{
		{
			name:     "the observed object form",
			in:       `{"scsi0":true}`,
			accepted: true,
			scsi0:    true,
		},
		{
			name:     "an object can report a rejection per key",
			in:       `{"scsi0":false}`,
			accepted: true,
			scsi0:    false,
		},
		{
			// The shape that used to fail to decode. Its sibling fields
			// in the same response are plain bools, so it is the shape
			// most likely to turn up.
			name:     "a bare true is an acceptance with nothing to enumerate",
			in:       `true`,
			accepted: true,
			scsi0:    false,
		},
		{
			name:     "a bare false is not an acceptance",
			in:       `false`,
			accepted: false,
		},
		{
			name:     "a quoted bool, as this API sends elsewhere",
			in:       `"1"`,
			accepted: true,
		},
		{
			name:     "null is not an acceptance",
			in:       `null`,
			accepted: false,
		},
		{
			name:     "an empty object is an acceptance",
			in:       `{}`,
			accepted: true,
		},
		{
			// A genuine surprise must surface rather than be swallowed;
			// silently decoding it would be the bug this type exists to
			// avoid, pointed a third way.
			name:    "a list is not a shape we know",
			in:      `["scsi0"]`,
			wantErr: true,
		},
		{
			name:    "a string that is not a bool is not a shape we know",
			in:      `"maybe"`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got shtypes.MaybeBoolMap
			err := json.Unmarshal([]byte(tc.in), &got)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Unmarshal(%s) = %#v, want an error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal(%s): %v", tc.in, err)
			}
			if got.Accepted() != tc.accepted {
				t.Errorf("Accepted() = %t, want %t", got.Accepted(), tc.accepted)
			}
			if got.AcceptedKey("scsi0") != tc.scsi0 {
				t.Errorf("AcceptedKey(scsi0) = %t, want %t", got.AcceptedKey("scsi0"), tc.scsi0)
			}
		})
	}
}

// TestMaybeBoolMap_InAResponse checks the field decodes in the position
// it actually occupies, alongside the plain bools it sits next to.
func TestMaybeBoolMap_InAResponse(t *testing.T) {
	t.Parallel()

	var resp struct {
		Return struct {
			Cores bool                 `json:"cores"`
			RAM   bool                 `json:"ram"`
			Disk  shtypes.MaybeBoolMap `json:"disk"`
		} `json:"return"`
	}

	// The scalar form, which used to fail outright.
	if err := json.Unmarshal([]byte(`{"return":{"cores":false,"ram":false,"disk":false}}`), &resp); err != nil {
		t.Fatalf("scalar disk must not fail the whole decode: %v", err)
	}
	if resp.Return.Disk.Accepted() {
		t.Error(`Disk.Accepted() = true for "disk": false`)
	}

	if err := json.Unmarshal([]byte(`{"return":{"cores":false,"ram":false,"disk":{"scsi0":true}}}`), &resp); err != nil {
		t.Fatalf("object disk: %v", err)
	}
	if !resp.Return.Disk.AcceptedKey("scsi0") {
		t.Error("AcceptedKey(scsi0) = false for the observed success shape")
	}
}
