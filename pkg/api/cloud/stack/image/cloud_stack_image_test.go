package image_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
)

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	got, err := image.New(ex.Client).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, image.ListResponse{})

	if len(got.Return) == 0 {
		t.Fatal("no images decoded; the catalogue is never empty")
	}

	var withVersions int
	for i, img := range got.Return {
		if img.Code == "" {
			t.Errorf("Return[%d].Code is empty; the code is how an image is referenced", i)
		}
		if len(img.Versions) > 0 {
			withVersions++
		}
	}
	if withVersions == 0 {
		t.Error("no image decoded with a version; versions are a list here, unlike the object nested under a container's image_details")
	}

	// Labels is an object on this endpoint. It is a JSON-encoded string
	// on stack image versions and inside image_details, so a single
	// shared type would be wrong; this asserts which one applies here.
	var withLabels int
	for _, img := range got.Return {
		if len(img.Labels) > 0 {
			withLabels++
		}
	}
	if withLabels == 0 {
		t.Error("no image decoded with labels; on this endpoint labels is an object, not a string")
	}
}
