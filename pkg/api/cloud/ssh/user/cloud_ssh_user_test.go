package user_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	user "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
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

	var withKeys, withContainers int
	for i, u := range got.Return.Users {
		if u.Username == "" {
			t.Errorf("Users[%d].Username is empty", i)
		}
		if u.HomeDir == "" {
			t.Errorf("Users[%d].HomeDir is empty", i)
		}
		if len(u.SSHKeys) > 0 {
			withKeys++
		}
		if len(u.Containers) > 0 {
			withContainers++
		}
	}

	// The fixture holds both cases, which is what makes these checks
	// able to fail: a user with keys and one without.
	if withKeys == 0 {
		t.Error("no user decoded with an SSH key; the key list is what this endpoint is for")
	}
	if withContainers == 0 {
		t.Error("no user decoded with a container; the grant list did not decode")
	}
}
