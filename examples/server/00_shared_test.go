package main

import (
	"reflect"
	"testing"
)

// TestDeletable_NeverTargetsServersItDidNotCreate pins the safety
// property that matters most in this example.
//
// The bug this replaces: deletable() fell back to SH_SERVER_A and
// SH_SERVER_B when nothing had been provisioned. runJourney calls
// cleanup unconditionally, including after a failed tour — correct, so
// a freshly provisioned server is not leaked — so a journey that failed
// before stepProvision recorded anything would delete whatever those
// two variables named. They are exactly what the README tells you to
// export for standalone runs, so that is the normal working state.
// Deletion is forced and not recoverable.
func TestDeletable_NeverTargetsServersItDidNotCreate(t *testing.T) {
	// Not parallel: these subtests set process environment.
	t.Run("names a read-only step was given are not consent to destroy", func(t *testing.T) {
		t.Setenv("SH_SERVER_A", "someones-production-server")
		t.Setenv("SH_SERVER_B", "another-one")
		t.Setenv("SH_DELETE_SERVERS", "")

		var st state
		if got := st.deletable(); len(got) != 0 {
			t.Fatalf("deletable() = %v, want nothing — SH_SERVER_A/B must never be deleted implicitly", got)
		}
	})

	t.Run("what this process created is cleaned up", func(t *testing.T) {
		t.Setenv("SH_SERVER_A", "someones-production-server")
		t.Setenv("SH_DELETE_SERVERS", "")

		st := state{created: []string{"gosh-journey-a", "gosh-journey-b"}}
		want := []string{"gosh-journey-a", "gosh-journey-b"}
		if got := st.deletable(); !reflect.DeepEqual(got, want) {
			t.Errorf("deletable() = %v, want %v — a provisioned server must not be leaked", got, want)
		}
	})

	t.Run("deleting something else takes its own opt-in", func(t *testing.T) {
		t.Setenv("SH_SERVER_A", "not-this-one")
		t.Setenv("SH_DELETE_SERVERS", "old-a, old-b ,")

		var st state
		want := []string{"old-a", "old-b"}
		if got := st.deletable(); !reflect.DeepEqual(got, want) {
			t.Errorf("deletable() = %v, want %v", got, want)
		}
	})

	t.Run("what this process created wins over the opt-in", func(t *testing.T) {
		t.Setenv("SH_DELETE_SERVERS", "something-else")

		st := state{created: []string{"gosh-journey-a"}}
		want := []string{"gosh-journey-a"}
		if got := st.deletable(); !reflect.DeepEqual(got, want) {
			t.Errorf("deletable() = %v, want %v", got, want)
		}
	})
}
