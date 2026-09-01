package shapecheck

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type inner struct {
	Name string `json:"name"`
}

type envelope struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
}

type paged struct {
	Return struct {
		Total int     `json:"total_items"`
		Data  []inner `json:"data"`
	} `json:"return"`
	envelope // embedded, promoted, no tag
}

// TestUndecoded_FindsAMissingField is the case the whole package exists
// for: a field the API sends that nothing decodes.
func TestUndecoded_FindsAMissingField(t *testing.T) {
	t.Parallel()
	body := `{"status":true,"msg":"ok","return":{"total_items":1,"data":[{"name":"x","container":"y"}]}}`

	got, err := Undecoded([]byte(body), paged{})
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	want := []string{"return.data[].container"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Undecoded = %v, want %v", got, want)
	}
}

// TestUndecoded_CleanTypeReportsNothing is the other half. Without it
// the package could return everything and still pass the test above.
func TestUndecoded_CleanTypeReportsNothing(t *testing.T) {
	t.Parallel()
	body := `{"status":true,"msg":"ok","return":{"total_items":1,"data":[{"name":"x"}]}}`

	got, err := Undecoded([]byte(body), paged{})
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v, want nothing — every field has a home", got)
	}
}

// TestUndecoded_FollowsEmbeddedStructs checks a promoted field is found
// on the parent, as encoding/json does. Missing this would report the
// whole shared envelope as undecoded on every single response type.
func TestUndecoded_FollowsEmbeddedStructs(t *testing.T) {
	t.Parallel()
	got, err := Undecoded([]byte(`{"status":true,"msg":"ok"}`), paged{})
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v; status and msg come from the embedded envelope", got)
	}
}

// TestUndecoded_MatchesUntaggedFieldsCaseInsensitively mirrors
// encoding/json. Several response types in this SDK declare Return with
// no tag and rely on that match.
func TestUndecoded_MatchesUntaggedFieldsCaseInsensitively(t *testing.T) {
	t.Parallel()
	var v struct {
		Return struct {
			Data []inner `json:"data"`
		}
	}
	got, err := Undecoded([]byte(`{"return":{"data":[{"name":"x"}]}}`), v)
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v; an untagged Return must match \"return\"", got)
	}
}

// TestUndecoded_TreatsCatchAllsAsCovering checks interface{},
// json.RawMessage and map fields absorb whatever is under them, which is
// what they are for. Container.ImageDetails depends on this.
func TestUndecoded_TreatsCatchAllsAsCovering(t *testing.T) {
	t.Parallel()
	var v struct {
		Any  any                    `json:"any"`
		Raw  json.RawMessage        `json:"raw"`
		Dict map[string]interface{} `json:"dict"`
	}
	body := `{"any":{"a":1},"raw":{"labels":"{}","versions":{"id":"1"}},"dict":{"x":{"y":2}}}`
	got, err := Undecoded([]byte(body), v)
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v; a catch-all field covers everything beneath it", got)
	}
}

// TestUndecoded_DeduplicatesAcrossListElements keeps the report
// readable: one line per path, not one per row.
func TestUndecoded_DeduplicatesAcrossListElements(t *testing.T) {
	t.Parallel()
	body := `{"return":{"total_items":3,"data":[{"name":"a","extra":1},{"name":"b","extra":2},{"name":"c","extra":3}]}}`
	got, err := Undecoded([]byte(body), paged{})
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 1 || got[0] != "return.data[].extra" {
		t.Errorf("Undecoded = %v, want one deduplicated path", got)
	}
}

// TestUndecoded_HandlesCustomUnmarshalers checks a shtypes field does
// not look like a struct to descend into.
func TestUndecoded_HandlesCustomUnmarshalers(t *testing.T) {
	t.Parallel()
	var v struct {
		Managed shtypes.MaybeBool `json:"managed"`
	}
	got, err := Undecoded([]byte(`{"managed":"0"}`), v)
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v, want nothing", got)
	}
}

// TestUndecoded_RejectsNonJSON checks a truncated body is an error, not
// a silent all-clear.
func TestUndecoded_RejectsNonJSON(t *testing.T) {
	t.Parallel()
	if _, err := Undecoded([]byte("<html>502</html>"), paged{}); err == nil {
		t.Fatal("expected an error for a non-JSON body")
	}
}

// TestUndecoded_AcceptsAPointer checks callers can pass either form.
func TestUndecoded_AcceptsAPointer(t *testing.T) {
	t.Parallel()
	got, err := Undecoded([]byte(`{"return":{"data":[{"name":"x","gone":1}]}}`), &paged{})
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("Undecoded = %v, want the one missing field", got)
	}
}

type concrete struct {
	A string `json:"a"`
}

// TestUndecoded_DescendsIntoTypedMapValues records the distinction
// between a map that absorbs anything and a map whose values have a
// type.
//
// map[string]interface{} is a genuine catch-all. map[string]Concrete is
// not: the keys are unknown but the values can be missing fields like
// any struct. Treating both as catch-alls meant three endpoints whose
// Return is a map of a concrete type — redirect.ListRedirects,
// server.ListAllocatedIPs, and the disk map on ListUpgrades — were
// reported clean whatever the API sent.
func TestUndecoded_DescendsIntoTypedMapValues(t *testing.T) {
	t.Parallel()

	var v struct {
		M map[string]concrete `json:"m"`
	}
	got, err := Undecoded([]byte(`{"m":{"a-key":{"a":"x","surprise":1}}}`), v)
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	want := []string{"m.a-key.surprise"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Undecoded = %v, want %v — a map of a concrete type is not a catch-all", got, want)
	}

	// And a genuine catch-all still absorbs, so the distinction is the
	// value type rather than the map.
	var anyMap struct {
		M map[string]interface{} `json:"m"`
	}
	got, err = Undecoded([]byte(`{"m":{"a-key":{"a":"x","surprise":1}}}`), anyMap)
	if err != nil {
		t.Fatalf("Undecoded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Undecoded = %v, want nothing for map[string]interface{}", got)
	}
}
