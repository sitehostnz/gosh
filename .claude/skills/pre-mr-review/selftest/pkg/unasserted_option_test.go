// This file is a fixture, not a test the repository runs. It exists so
// premr.py --selftest can prove the unasserted-option check still
// fires, and it is deliberately the exact shape of the finding that
// check was written for: a request option set in the literal and never
// looked for in the handler.
//
// The check shipped unable to fire for any input at all, twice over —
// first because it searched the whole function body, which always
// contains the setter, and then because slicing from httptest.NewServer
// to the end of the function still contained it. Neither was visible in
// its output, because a check that cannot fire and a clean tree look
// identical. That is what this fixture is for.
package selftest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFixtureUnassertedOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Asserts Type and nothing else. SortBy and SortDir below are
		// sent and never checked, so this test would pass unchanged if
		// the client stopped sending them.
		if got := r.URL.Query().Get("filters[type]"); got != "hpvm-distro" {
			t.Errorf("type = %q", got)
		}
	}))
	defer srv.Close()

	_ = ListImagesOptions{
		Type:    "hpvm-distro",
		SortBy:  "name",
		SortDir: "desc",
	}
}
