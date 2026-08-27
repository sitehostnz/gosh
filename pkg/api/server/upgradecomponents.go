package server

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpgradeComponents upgrades specific hardware components — cores, RAM
// and/or disk — on a server via /server/upgrade.json. Returns which
// components were accepted, and a job when the work happens out of
// band.
//
// At least one of Cores, RAM or Disk should be set; passing nothing is
// accepted by the API but is a no-op.
//
// # What each product accepts
//
// Verified live, August 2026:
//
//	                      cores / RAM              disk
//	High performance      rejected                 accepted
//	Standard performance  accepted                 accepted
//	Cloud Container       rejected                 (not tested)
//
// On high-performance products cores and RAM are fixed by the product
// code and this endpoint rejects them with "Please specify a valid
// cores value." / "Please specify a valid ram value." — change them
// with Upgrade (plan change) instead. Disk is different: it grows
// happily on high performance, so this endpoint is not simply
// unavailable there.
//
// Naming note: this wraps /server/upgrade.json (component upgrade).
// The existing server.Upgrade method historically wraps
// /server/upgrade_plan.json (plan / product-code upgrade) — its name
// predates this wrapper, retained for backwards compatibility.
//
// An earlier note here said component upgrades were rejected by
// high-performance and CCS products outright, generalising from a
// cores/RAM test on CLDCON4-P. That is too broad: disk upgrades work on
// high performance. Corrected August 2026.
// Disk growth behaves differently by platform, verified live in
// August 2026:
//
//   - **High performance (HPVS).** The resize is applied online and
//     immediately. The server reports the new Size straight away,
//     NewSize stays zero, no reboot is needed and there is no job to
//     poll. CommitDiskChanges is not required.
//   - **Standard performance.** The intended size is staged as the
//     partition's NewSize and CommitDiskChanges applies it.
//
// A caller that wants to work on both should read the partition back
// after this call: if NewSize is set, commit it; if Size already
// reflects the request, the resize is done. examples/server does this.
//
// The platform may need to migrate a server to another node to find
// space, in which case the operation takes considerably longer than a
// local resize.
func (s *Client) UpgradeComponents(ctx context.Context, request UpgradeComponentsRequest) (response UpgradeComponentsResponse, err error) {
	u := "server/upgrade.json"
	keys := []string{"client_id", "name"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)
	if request.Cores > 0 {
		values.Add("upgrade[cores]", strconv.Itoa(request.Cores))
		keys = append(keys, "upgrade[cores]")
	}
	if request.RAM != "" {
		values.Add("upgrade[ram]", request.RAM)
		keys = append(keys, "upgrade[ram]")
	}
	for label, sz := range request.Disk {
		k := "upgrade[disk][" + label + "]"
		values.Add(k, strconv.Itoa(sz))
		keys = append(keys, k)
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
