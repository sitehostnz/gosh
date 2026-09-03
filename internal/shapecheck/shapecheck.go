// Package shapecheck compares a recorded API response against the Go
// type that decodes it, and reports what the type does not account for.
//
// # The gap it closes
//
// encoding/json ignores what it does not recognise. A field the API
// sends that no struct field claims is dropped in silence, so a decode
// can succeed completely while losing half the response — and a test
// that checks the fields it does know about will pass while it happens.
//
// That is not hypothetical in this SDK. cloud/db returned a container
// field that models.Database never declared; nothing failed, the value
// just never arrived. Undecoded is the failure mode that has no symptom
// until someone needs the value.
//
// So the check runs the other way round from a normal assertion: rather
// than asking whether the fields we expect are present, it asks whether
// anything present is unexpected.
package shapecheck

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Undecoded returns the JSON paths in body that v has no field for,
// sorted. An empty result means the type accounts for the whole
// response.
//
// v may be a struct or a pointer to one; embedded structs are followed,
// as are slices and maps, so a field buried in a paged list is reported
// with its full path ("return.data[].container").
func Undecoded(body []byte, v any) ([]string, error) {
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("shapecheck: %w", err)
	}
	var out []string
	walk("", doc, reflect.TypeOf(v), &out)
	return dedupe(out), nil
}

// walk descends the decoded document and the type together.
func walk(path string, doc any, t reflect.Type, out *[]string) {
	if t == nil {
		return
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch d := doc.(type) {
	case map[string]any:
		walkObject(path, d, t, out)
	case []any:
		walkArray(path, d, t, out)
	}
}

// walkObject matches an object's keys against a struct's fields.
func walkObject(path string, d map[string]any, t reflect.Type, out *[]string) {
	// An interface{} field swallows anything, by design.
	if t.Kind() == reflect.Interface {
		return
	}

	// A map's keys are unknown, but its values have a type, and that
	// type can be incomplete like any other. Returning here treated
	// every map as a catch-all, so three endpoints whose Return is a
	// map of a concrete struct — redirect.ListRedirects,
	// server.ListAllocatedIPs, and the disk map on ListUpgrades — got a
	// clean bill of health whatever the API sent.
	//
	// That is the opposite of what this package is for: a silent
	// all-clear is the failure it exists to eliminate, and it was
	// issuing one.
	//
	// Reporting through the real key gives a better path too —
	// return.example.co.nz.destination rather than return[].destination.
	if t.Kind() == reflect.Map {
		for k, v := range d {
			walk(join(path, k), v, t.Elem(), out)
		}
		return
	}
	if t.Kind() != reflect.Struct {
		return
	}
	fields := jsonFields(t)
	for k, v := range d {
		ft, ok := lookup(fields, k)
		if !ok {
			*out = append(*out, join(path, k))
			continue
		}
		walk(join(path, k), v, ft, out)
	}
}

// walkArray descends into each element against the slice's element type.
func walkArray(path string, d []any, t reflect.Type, out *[]string) {
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
		return
	}
	for _, el := range d {
		walk(path+"[]", el, t.Elem(), out)
	}
}

// jsonFields maps a struct's wire names to their types, following
// embedded structs so that a promoted field is found under the parent.
func jsonFields(t reflect.Type) map[string]reflect.Type {
	out := make(map[string]reflect.Type)
	for i := range t.NumField() {
		f := t.Field(i)
		// encoding/json cannot write to an unexported field, so one
		// covers nothing. Counting it here would report a dropped
		// field as decoded — a false negative in the one tool whose
		// entire value is not producing them. The Anonymous case
		// stays: an embedded unexported type still promotes its
		// exported fields.
		if !f.IsExported() && !f.Anonymous {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			if promoted, ok := promotedFields(f.Type); ok {
				for k, v := range promoted {
					out[k] = v
				}
				continue
			}
		}
		if name == "" {
			name = f.Name
		}
		out[name] = f.Type
	}
	return out
}

// lookup finds a field by wire name, falling back to a case-insensitive
// match because encoding/json does the same. A field declared without a
// json tag is matched by its Go name, so "return" finds Return.
func lookup(fields map[string]reflect.Type, key string) (reflect.Type, bool) {
	if t, ok := fields[key]; ok {
		return t, true
	}
	for k, t := range fields {
		if strings.EqualFold(k, key) {
			return t, true
		}
	}
	return nil, false
}

// dedupe collapses the repeated paths that come of walking every
// element of a list, and sorts what is left. One report per path is
// what a reader needs; ninety copies of it is not.
func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// promotedFields returns the fields an embedded struct contributes to
// its parent, as encoding/json promotes them.
func promotedFields(t reflect.Type) (map[string]reflect.Type, bool) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, false
	}
	return jsonFields(t), true
}

// join builds a dotted path.
func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
