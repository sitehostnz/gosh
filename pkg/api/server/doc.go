// Package server wraps the SiteHost /server API endpoints — the
// VPS / Cloud Container Server lifecycle: provision, get, list,
// upgrade, snapshot, etc.
//
// # IP allocation when provisioning a new server
//
// Provisioning via Create requires an IPv4 (and optionally IPv6).
// There are three paths consumers should know about — the API
// docs only fully describe the first, so this package-level note
// captures all three for AI agents and humans reading the SDK:
//
//  1. **Auto-allocation (recommended for most cases).** Set
//     CreateRequest.Params.IPv4 to []string{"auto"} (and likewise
//     IPv6 if you want one). The platform picks a free address
//     from the location's pool and binds it to the new server.
//     The public docs explicitly recommend this:
//     "simply pass the string 'auto' to automatically assign
//     an IPv4 address."
//
//  2. **Specific pre-allocated address.** Pass the address(es)
//     directly, e.g. []string{"203.0.113.10"}. The address must
//     **already be allocated to the calling client_id** — the
//     platform won't transfer pool IPs into your client at
//     provision time via this path.
//
//     ListIPs(location) returns the IPs **currently allocated to
//     this client** at that location — *not* the location's free
//     pool. If ListIPs returns an empty slice, that does **not**
//     mean the pool is exhausted; it means this client has no
//     allocations there. Use ListLocations to read pool-wide
//     capacity (`AvailableIPs`, `AvailableIPv4`, `AvailableIPv6`).
//
//     Pitfall: a previous gosh session burned ~30 minutes
//     retrying provisions because ListIPs returned [] and the
//     wrapper's error message implied "no free IPs in the pool"
//     — but the pool had hundreds free; the right fix was to
//     pass `auto` instead. Don't waste cycles re-discovering
//     this.
//
//  3. **Manual allocation by SiteHost staff.** Reseller-style
//     arrangements may have IPs reserved for a client by
//     SiteHost ops; once allocated they're visible via ListIPs
//     and can be passed via path (2). Out of band relative to
//     the SDK.
//
// **Default to path (1) unless you have a reason not to.** If you
// do need a specific address, sanity-check via ListIPs first; if
// that returns empty, fall back to "auto" rather than retrying or
// concluding the pool is dry.
package server
