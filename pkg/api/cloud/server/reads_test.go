package server_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	server "github.com/sitehostnz/gosh/pkg/api/cloud/server"
)

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	got, err := server.New(ex.Client).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, server.ListResponse{})

	if len(got.CloudServers) == 0 {
		t.Fatal("no servers decoded; the fixture has rows")
	}
	for i, s := range got.CloudServers {
		if s.Name == "" {
			t.Errorf("CloudServers[%d].Name is empty; the name identifies the server everywhere else", i)
		}
		// These three were being dropped until a recorded response was
		// compared against the type.
		if s.Created == "" {
			t.Errorf("CloudServers[%d].Created is empty", i)
		}
		if s.DateUpdated == "" {
			t.Errorf("CloudServers[%d].DateUpdated is empty", i)
		}
	}
}

// TestGetUpdateWindow_RejectsAnUnmanagedServer records that update
// windows are a managed-service feature. The API answers HTTP 200 with
// status:false, so this also checks the envelope is what decides.
func TestGetUpdateWindow_RejectsAnUnmanagedServer(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "get_update_window-unmanaged.json")

	_, err := server.New(ex.Client).GetUpdateWindow(context.Background(),
		server.GetUpdateWindowRequest{ServerName: "s"})
	if err == nil {
		t.Fatal("GetUpdateWindow: expected an error for an unmanaged server")
	}
}
