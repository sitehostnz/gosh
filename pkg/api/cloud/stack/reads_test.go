package stack_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
)

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	got, err := stack.New(ex.Client).List(context.Background(), stack.ListRequest{ServerName: "s"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, stack.ListResponse{})

	if len(got.Return.Stacks) == 0 {
		t.Fatal("no stacks decoded; the fixture has rows")
	}

	var withContainers int
	for i, s := range got.Return.Stacks {
		if s.Name == "" {
			t.Errorf("Stacks[%d].Name is empty; the name is what every other endpoint wants", i)
		}
		if len(s.Containers) > 0 {
			withContainers++
		}
	}
	if withContainers == 0 {
		t.Error("no stack decoded with a container")
	}

	if got := ex.Request.URL.Query().Get("filters[server_name]"); got != "s" {
		t.Errorf("filters[server_name] = %q, want s", got)
	}
}

// TestGet_UsesServerNotServerName pins the one endpoint in this package
// that names the server differently. Sending server_name is rejected
// with "The server name is missing.", which reads like an omission
// rather than a misnaming, so it is worth a test that says why.
func TestGet_UsesServerNotServerName(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "get.json")

	got, err := stack.New(ex.Client).Get(context.Background(), stack.GetRequest{
		ServerName: "s", Name: "infra",
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, stack.GetResponse{})

	q := ex.Request.URL.Query()
	if q.Get("server") != "s" {
		t.Errorf("server = %q, want s — this endpoint does not accept server_name", q.Get("server"))
	}
	if q.Get("server_name") != "" {
		t.Error("server_name is on the query; this endpoint ignores it and then reports the name as missing")
	}
	if q.Get("name") != "infra" {
		t.Errorf("name = %q, want infra", q.Get("name"))
	}

	if got.Stack.DockerFile == "" {
		t.Error("DockerFile is empty; the compose file is the substance of a stack")
	}
}

// TestGet_ContainerImageDetailsStayRaw guards the reason
// Container.ImageDetails is a json.RawMessage: its labels field is a
// JSON-encoded string and its versions field is an object, neither of
// which agrees with models.StackImage.
func TestGet_ContainerImageDetailsStayRaw(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "get.json")

	got, err := stack.New(ex.Client).Get(context.Background(), stack.GetRequest{ServerName: "s", Name: "n"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Stack.Containers) == 0 {
		t.Fatal("no containers decoded")
	}

	var withDetails int
	for _, c := range got.Stack.Containers {
		if len(c.ImageDetails) > 0 {
			withDetails++
		}
	}
	if withDetails == 0 {
		t.Error("no container carried image_details; the raw field is not being populated")
	}
}
