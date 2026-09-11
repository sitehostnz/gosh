// This file is a fixture, not compiled code. It exists so
// premr.py --selftest proves the wire-contract check still recognises a
// parameter name it cannot read.
//
// Both shapes below defeated the check by the same mechanism: a quoted
// literal adjacent to "+" is a fragment, and extracting the fragment
// looked like success. Because literals were found, the dynamic flag
// stayed clear, so the check reported these functions clean having
// never seen a single real parameter name.
//
// The adversarial input is the point. A fixture built only from the
// happy-path bug — a key plainly missing from a plain keys list —
// passes against a check that cannot read half the call sites in the
// repository.
package selftest

import "github.com/sitehostnz/gosh/pkg/net"

// AddWithConcatenatedKey is the shape from cloud/stack/add.go: literal
// first, variable in the middle.
func AddWithConcatenatedKey(name string) {
	keys := []string{"client_id"}
	values := newValues()
	values.Add("client_id", "1")
	keys = append(keys, "environments["+name+".env]")
	values.Add("environments["+name+".env]", "vars")
	_ = net.Encode(values, keys)
}

// UpdateWithPrefixedKey is the shape from securitygroups/update.go:
// variable first, literal second — invisible to an extractor that
// requires the literal to lead.
func UpdateWithPrefixedKey(prefix string) {
	keys := []string{"client_id"}
	values := newValues()
	values.Add("client_id", "1")
	keys = append(keys, prefix+"[enabled]")
	values.Add(prefix+"[enabled]", "1")
	_ = net.Encode(values, keys)
}
