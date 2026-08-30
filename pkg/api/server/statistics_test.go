package server

import (
	"encoding/json"
	"testing"
)

// TestStatisticTypes_DecodesBothShapes covers the reason this endpoint
// had never once returned a metric name.
//
// A server with no metrics answers "[]"; a server with some answers an
// object. The field was declared []string, which decoded the first and
// failed the second — so it worked only where there was nothing to
// report, and the failure was invisible because the servers it was
// tried against were all of the first kind.
func TestStatisticTypes_DecodesBothShapes(t *testing.T) {
	t.Parallel()

	t.Run("a server with no metrics answers with an empty list", func(t *testing.T) {
		t.Parallel()
		var got StatisticTypes
		if err := json.Unmarshal([]byte(`[]`), &got); err != nil {
			t.Fatalf("the empty-list form must decode: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("want an empty map rather than nil, so a caller can range over it safely")
		}
	})

	t.Run("a server with metrics answers with an object", func(t *testing.T) {
		t.Parallel()
		// Taken from a live Xen server. Note XenCpu's parameter list
		// holds an empty element expressed as "[]", not "{}" — PHP
		// serialising an empty map as a list.
		const body = `{
			"XenCpu": [[]],
			"XenDiskIO": [{"partition":"a-disk"},{"partition":"a-swap"}],
			"XenNetwork": [{"iface":"e0"}]
		}`
		var got StatisticTypes
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatalf("the object form must decode: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
		if n := len(got["XenDiskIO"]); n != 2 {
			t.Errorf("XenDiskIO has %d parameter set(s), want 2", n)
		}
		if got["XenDiskIO"][0].Partition != "a-disk" {
			t.Errorf("XenDiskIO[0].Partition = %q, want a-disk", got["XenDiskIO"][0].Partition)
		}
		if got["XenNetwork"][0].Iface != "e0" {
			t.Errorf("XenNetwork[0].Iface = %q, want e0", got["XenNetwork"][0].Iface)
		}
		// The empty element must survive rather than fail the decode:
		// it means "this metric takes no item".
		if n := len(got["XenCpu"]); n != 1 {
			t.Fatalf("XenCpu has %d element(s), want 1", n)
		}
		if got["XenCpu"][0].Partition != "" || got["XenCpu"][0].Iface != "" {
			t.Errorf("XenCpu[0] = %+v, want an empty parameter set", got["XenCpu"][0])
		}
	})

	t.Run("a shape we do not understand is an error", func(t *testing.T) {
		t.Parallel()
		var got StatisticTypes
		if err := json.Unmarshal([]byte(`"XenCpu"`), &got); err == nil {
			t.Errorf("Unmarshal = %#v, want an error rather than a silently empty value", got)
		}
	})
}

// TestGetStatistics_RequiresType pins the parameter whose absence made
// this endpoint unusable. It was missing from the options struct
// entirely, so every call came back with the type reported as missing.
func TestGetStatistics_RequiresType(t *testing.T) {
	t.Parallel()

	c := New(nil)
	if _, err := c.GetStatistics(nil, GetStatisticsOptions{ServerName: "s"}); err == nil { //nolint:staticcheck // the guard returns before the context is used
		t.Fatal("expected an error when Type is empty")
	}
	if _, err := c.GetStatistics(nil, GetStatisticsOptions{Type: "XenCpu"}); err == nil { //nolint:staticcheck // as above
		t.Fatal("expected an error when ServerName is empty")
	}
}
