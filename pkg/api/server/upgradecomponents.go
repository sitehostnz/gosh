package server

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpgradeComponents upgrades specific hardware components (cores
// and/or RAM) on a server via /server/upgrade.json. Returns a
// scheduler job plus per-component bool flags indicating which
// upgrades were accepted.
//
// At least one of Cores or RAM should be set; passing zero for both
// is accepted by the API but is a no-op.
//
// Naming note: this wraps /server/upgrade.json (component upgrade).
// The existing server.Upgrade method historically wraps
// /server/upgrade_plan.json (plan / product-code upgrade) — its name
// predates this wrapper, retained for backwards compatibility.
//
// **Live finding** (May 2026): the API rejects component upgrades
// against CCS products (CLDCON4-P tested) with "Please specify a
// valid cores value." — these products have fixed cores/RAM tied
// to the product code; component-level scaling appears to be a
// VPS-only operation. Use server.Upgrade (plan / product-code
// upgrade) for CCS resizing instead.
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
