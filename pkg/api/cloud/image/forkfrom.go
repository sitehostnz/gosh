package image

import (
	"context"
	"fmt"
	"strconv"
)

// ForkFromImage creates a new custom image forked from a SiteHost
// public image identified by parentCode (e.g. "sitehost-php80").
//
// This composite helper resolves the public parent's numeric id by
// listing images and matching on Code+IsPublic, then calls Create
// with that fork_id. Without this helper, consumers have to do the
// list-filter-extract-id dance themselves.
//
// label is required (used as the new image's display label).
// code is the desired image code (slug used in the GitLab repo URL
// and registry); pass "" to let the API auto-generate it from label.
// sshKeyIDs are customer-level SSH key IDs that get push access to
// the backing GitLab repository — at least one is required for the
// consumer to push commits.
//
// Returns the JobResponse from Create. Consumers should poll the
// job to completion before calling CloneURL or attempting a clone,
// since the GitLab repo is provisioned asynchronously.
func (s *Client) ForkFromImage(ctx context.Context, parentCode, label, code string, sshKeyIDs []int) (response JobResponse, err error) {
	if parentCode == "" {
		return response, fmt.Errorf("cloud.image.ForkFromImage: parentCode is required")
	}
	if label == "" {
		return response, fmt.Errorf("cloud.image.ForkFromImage: label is required")
	}

	listing, err := s.List(ctx)
	if err != nil {
		return response, fmt.Errorf("listing images to resolve fork target %q: %w", parentCode, err)
	}

	var forkID int
	for _, img := range listing.Return.Images {
		if img.Code == parentCode && bool(img.IsPublic) {
			id, perr := strconv.Atoi(img.ID)
			if perr != nil {
				return response, fmt.Errorf("public image %q has non-numeric id %q: %w", parentCode, img.ID, perr)
			}
			forkID = id
			break
		}
	}
	if forkID == 0 {
		return response, fmt.Errorf("public SiteHost image with code %q not found in cloud.image.list_all", parentCode)
	}

	return s.Create(ctx, CreateRequest{
		Label:   label,
		Code:    code,
		ForkID:  forkID,
		SSHKeys: sshKeyIDs,
	})
}
