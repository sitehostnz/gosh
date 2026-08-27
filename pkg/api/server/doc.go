// Package server wraps the SiteHost /server API endpoints — the
// VPS / Cloud Container Server lifecycle: provision, get, list,
// upgrade, snapshot, etc.
//
// # What you need before you can provision
//
// Most of what a provisioning flow needs is discoverable through this
// package. Two things are not, and they are the usual reason a first
// attempt fails.
//
//  1. **Location** — discoverable. ListLocations returns the code, the
//     product families the location carries (ProductTypes) and current
//     AvailableIPv4. A location with no free IPv4 cannot take a new
//     server.
//
//  2. **Product code** — discoverable, via [Client.ListProducts]. That
//     endpoint is undocumented rather than absent; see "Product codes"
//     below.
//
//  3. **Image** — discoverable, but the catalogue is split and the
//     default listing is not the one you want for a high-performance
//     product. See ListImages: high-performance images live behind
//     ImageTypeHPVMDistro plus a mandatory Location, and their codes
//     carry a build date, so they must be looked up rather than
//     pinned. A code from the wrong catalogue is rejected with
//     "The image '<code>' could not be found".
//
//     CanProvision does NOT validate the image. It can return success
//     for a code that Create then rejects.
//
//  4. **SSH key** — must be supplied at provision time. A key cannot
//     be injected afterwards, and the password in CreateResponse is
//     returned once and never again, so automation that misses it has
//     no way in.
//
//     Two steps, and both are required. Register the key on the
//     account first through the ssh/key endpoints, then pass its
//     **public key content** in ParamsOptions.SSHKeys — not the key id
//     that registration returned. Provisioning takes key material but
//     validates it against the account's registered keys, so an
//     unregistered key is rejected:
//
//     "One or more of the given SSH keys are not present in our
//     system. Please add them to your account, before attempting to
//     use them during a provision."
//
//     Passing the id instead of the content fails the same way, since
//     the id is not key material.
//
//  5. **The login username** — NOT exposed by the API. The platform
//     records a default login user per image and uses it when
//     provisioning, but ListImages does not return it, so a client has
//     to know it out of band. See LoginUserFor, which records the
//     mapping — it depends on the product family as well as the distro —
//     and what was not verified, notably Windows (reached over RDP
//     rather than SSH) and Private Cloud locations.
//
// # Product codes
//
// Use [Client.ListProducts]. It returns every product orderable at a
// location, with cores, RAM, disk, bandwidth, price and the disk
// labels — so nothing here needs to be hardcoded.
//
// It is worth saying why that is not obvious: the endpoint is
// **undocumented**. It does not appear in the public API documentation
// or its endpoint listing, which is why the Knowledge Base product-code
// page is maintained by hand and covers only the older VPS families.
// The endpoint is nonetheless public and supported.
//
// Products are scoped to a location's product group, so Location is
// required and there is no "list everything" call. That scoping also
// explains why the same code can carry different specifications in
// different places — bandwidth in particular varies. Asked per
// location, a code has exactly one specification.
//
// The three virtual-server tiers, since the naming does not make the
// ordering obvious:
//
//	family   tier                        example codes
//	LINVPS   Xen — legacy                XENLIT, XENPRO
//	SVS      standard performance (KVM)  LSVSP1–6, WSVS2–16
//	HPVS     high performance (KVM)      LHPVS1–30
//
// Filter with ListProductsOptions.Types using the ProductType*
// constants. Legacy Xen is offered in New Zealand and Australia only;
// FRA1 and USCAL1 are KVM-only sites, which is a reason to prefer HPVS
// for anything that may need to run outside those two countries.
//
// # Validating a product code
//
// ListProducts says what exists at a location. CanProvision says
// whether one can be placed there right now, and its three outcomes
// mean different things:
//
//	Successful                  a node can take this plan here, now
//	"Products not found"         the code is not in this location's
//	                             product group — genuinely not offered
//	                             here
//	"No available nodes found"   the product IS offered here, but no
//	                             node can currently fit it
//
// The distinction matters because the recovery differs. "Products not
// found" means the code is wrong for that location and no amount of
// retrying helps. "No available nodes found" means the plan is real and
// offered but unplaceable right now — either the location is full, or
// the plan is too large for the free headroom, which is why the biggest
// plans can fail where smaller ones succeed. Retry, choose a smaller
// plan, or choose another location.
//
// # Why CanProvision is the right way to ask
//
// Create does **not** perform the placement check. The request is
// accepted, a subscription is added and a provisioning job is queued;
// node selection happens later, when the job runs. If nothing can take
// the server the job then fails with "Not Enough Resources Available To
// Do This Right Now", leaving a server record and a subscription behind
// against a failed job.
//
// So attempting a provision is not a safe way to probe availability —
// a refusal costs nothing, but an acceptance that later fails leaves
// state someone has to clean up. Ask CanProvision first, and treat
// acceptance by Create as "queued", not as "capacity confirmed".
//
// # The server's name is not the label you sent
//
// Create takes a Label. The platform derives the server's Name from
// it, truncating it and appending a digit if it collides with an
// existing server — so labels "web-a" and "web-b" can become names
// "web" and "web1".
//
// Every other call in this package identifies a server by Name. Always
// read CreateResponse.Return.Name and use that. Assuming the name
// matches the label is the single easiest way to operate on the wrong
// server.
//
// # Rate limiting
//
// The API rate-limits per reseller — that is, per API key's owning
// reseller, not per key. The default allowance is 10 requests per
// second; it is configurable per reseller, so a given key may have more
// or less. Exceeding it returns HTTP 500 with "You have exceeded the
// number of requests per second for this key".
//
// That response is safe to retry, including for writes. The limit is
// checked after the key is authenticated but **before the request is
// dispatched**, so a throttled call never reaches the handler and
// cannot have had any effect — a rate-limited provision has not created
// anything.
//
// Job polling in a tight loop is the usual way to trip it.
//
// # Deleting
//
// A server cannot be deleted while it is still building: Delete is
// rejected with "The specified server cannot be deleted while in the
// 'Provisioning' state". Poll GetState and retry rather than treating
// that as a permanent failure, or the server leaks. Set
// DeleteRequest.Force for anything carrying containers.
//
// # Changing a server's addresses
//
// AddIP, RemoveIP and SetPrimaryIP have ordering rules that are not
// obvious and are documented on those methods. In short: a server will
// not accept an address from one network while it still holds one from
// another, so an address swap releases both servers before assigning
// either; and the guest's own network configuration is not updated by
// any of this — see GenerateNetworkConfig. examples/server walks the
// whole sequence.
//
// # Firewall
//
// The firewall subpackages apply to high-performance products only.
// See pkg/api/server/firewall.
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
