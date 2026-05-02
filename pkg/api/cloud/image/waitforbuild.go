package image

import (
	"context"
	"fmt"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/cloud/image/version"
)

// BuildStatus values returned by the platform CI for a custom image
// version. Compared as strings; not exhaustive — surface unknowns
// to the caller verbatim.
const (
	BuildStatusSuccess = "success"
	BuildStatusFailed  = "failed"
	BuildStatusRunning = "running"
)

// WaitForBuild polls cloud/image/version/list_all for the named
// image (by numeric image_id) until the most-recent version reports
// a terminal build status (success or failed), or the timeout is
// reached. interval controls poll cadence.
//
// Returns the terminating Version. If the build failed, the caller
// should fetch the trace via cloud.image.version.GetBuild(code,
// build_id) to surface the failure to the user.
//
// Why a helper: the API doesn't expose a build-watching endpoint,
// so consumers always have to poll. Doing it consistently here also
// protects against the platform's "Only the last successfully built
// version is available to be deployed" rule — we always look at the
// latest version, never an older success.
func (s *Client) WaitForBuild(ctx context.Context, imageID int, timeout, interval time.Duration) (version.Version, error) {
	if imageID == 0 {
		return version.Version{}, fmt.Errorf("cloud.image.WaitForBuild: imageID is required")
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}

	versionClient := version.New(s.client)
	deadline := time.Now().Add(timeout)

	for {
		resp, err := versionClient.ListAll(ctx, version.ListAllRequest{
			ImageID:  imageID,
			SortBy:   "date_added",
			SortDir:  "DESC",
			PageSize: 1,
		})
		if err != nil {
			return version.Version{}, fmt.Errorf("listing versions for image %d: %w", imageID, err)
		}

		if len(resp.Return.Versions) > 0 {
			latest := resp.Return.Versions[0]
			switch latest.BuildStatus {
			case BuildStatusSuccess, BuildStatusFailed:
				return latest, nil
			}
		}

		if time.Now().After(deadline) {
			latestStatus := "no versions yet"
			if len(resp.Return.Versions) > 0 {
				latestStatus = resp.Return.Versions[0].BuildStatus
			}
			return version.Version{}, fmt.Errorf("timed out after %s waiting for image %d build (last status: %s)",
				timeout, imageID, latestStatus)
		}

		select {
		case <-ctx.Done():
			return version.Version{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}
