package user_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	user "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
)

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	got, err := user.New(ex.Client).List(context.Background(), user.ListOptions{ServerName: "s"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, user.ListResponse{})

	if len(got.Return.Users) == 0 {
		t.Fatal("no users decoded; the fixture has rows")
	}

	// Passwords are write-only on this API. A non-empty one here would
	// mean either the API changed or the fixture was hand-written.
	for i, u := range got.Return.Users {
		if u.Password != "" {
			t.Errorf("Users[%d].Password is non-empty; listings never return one", i)
		}
		if u.Username == "" {
			t.Errorf("Users[%d].Username is empty", i)
		}
	}

	if got := ex.Request.URL.Query().Get("filters[server_name]"); got != "s" {
		t.Errorf("filters[server_name] = %q, want s", got)
	}
}

// TestList_FiltersAreOptional guards a fact worth not re-learning: the
// server name is a filter, not a requirement. Calling without one
// returns every database user on the account.
func TestList_FiltersAreOptional(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	if _, err := user.New(ex.Client).List(context.Background(), user.ListOptions{}); err != nil {
		t.Fatalf("List with no options: %v", err)
	}
	if got := ex.Request.URL.Query().Get("filters[server_name]"); got != "" {
		t.Errorf("filters[server_name] = %q; an unset filter must not reach the wire", got)
	}
}
