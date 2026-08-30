# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `server.ListImagesOptions` gives `server.ListImages` the filters the
  API already supported: `Type`, `Location`, `IncludeDisabled`,
  `PageSize`/`PageNumber` and `SortBy`/`SortDir`. This makes the
  high-performance (HPVS) image catalogue reachable for the first
  time — it sits behind `Type: server.ImageTypeHPVMDistro` plus a
  mandatory `Location`, and every code in the default listing is
  rejected by HPVS product codes. HPVS image codes carry a build date
  (`ubuntu-2404-20260727`), so they must be discovered, not hardcoded.
- `server.ImageTypeDistro`, `ImageTypeHPVMDistro`, `ImageTypeContainer`
  and `ImageTypeApp` name the closed set the type filter accepts.
- `server.Location*` constants for the location codes known at the time
  of writing. `server.ListLocations` remains authoritative.
- `server.ListProducts` wraps `server/products.json`, which lists every
  product orderable at a location with its cores, RAM, disk, bandwidth,
  price and disk labels. The endpoint is undocumented rather than
  absent — it is missing from the public API docs and their endpoint
  listing, though the platform changelog records it as public — which is
  why product codes have had to be hardcoded until now. `Location` is
  required, because products are scoped to a location's product group;
  `Types` and `Codes` filter the result.
  Two wire shapes are handled: `attributes` arrives as `[]` rather than
  `{}` for some products, and `partitions[].size` arrives as either a
  number or a string within one response. Attributes with no typed field
  are kept in `ProductAttributes.Extra` rather than dropped.
- `server.LoginUserFor` returns the account to log in as, which the API
  does not expose anywhere. It depends on the product family as well as
  the distro: the same Ubuntu image is `ubuntu` on high-performance and
  `root` on standard-performance. Determined empirically; the doc
  comment records what was not verified.
- `server.ProductTypeHPVS`, `ProductTypeSVS`, `ProductTypeLINVPS` and
  `ProductTypeWINVPS` name the product families
  `models.Server.ProductType` reports. `ProductTypeSVS` matters most:
  it is the family `server.LoginUserFor` deliberately does not cover,
  so a caller needs the constant in order to write that check.
- `examples/server`: the server lifecycle as a numbered journey —
  register a key, discover an image, provision a pair, optionally
  resize or change plan, swap their addresses, complete the cutover
  inside the guests, delete everything. The file numbering encodes the
  API's ordering constraints, which are otherwise undocumented. Every
  step runs standalone against existing servers as well as in sequence.
  Provisions real servers, so it does nothing without
  `SH_EXAMPLE_ALLOW_PROVISION=1`; run with no arguments it prints the
  journey map and exits zero.

### Fixed

- `examples/server`: the `delete` step deleted servers it did not
  create. When nothing had been provisioned it fell back to
  `SH_SERVER_A`/`SH_SERVER_B` — the variables the README tells you to
  export for standalone runs — and the journey runs cleanup even after
  a failed tour, so a run that failed before provisioning would
  force-delete two servers the operator already owned. Deleting a
  server this process did not create now requires `SH_DELETE_SERVERS`,
  which exists for no other purpose.
- `examples/server`: `SH_BASE_URL` was documented in two places and
  read nowhere, so pointing the journey at a sandbox silently ran it
  against production.
- `examples/server`: `SH_SSH_KEY_FILE` was unreachable — steps 50 and
  80 rejected before consulting it, so following the error message's
  own advice produced the identical error.
- `examples/server`: a failure mid-swap left a server holding no
  address with no rollback. Released addresses are now restored on the
  way out, best-effort and loudly.
- `server.ProductAttributes` retained a typed field in `Extra` when the
  API spelled the key with different case, since `encoding/json`
  matches case-insensitively but the cleanup did not.


- `server.ListUpgrades` decodes at all. `QuotaUsage.Total` and `Used`
  were `int` where the API sends a fractional RAM quota (67.5 GB
  observed live), so the call failed with "cannot unmarshal number 67.5
  into ... of type int" — the endpoint had never worked. They are now
  `float64`. The test fixture that enshrined integer quotas is corrected
  against the wire.
- `server.Upgrades` gained the three fields the response carries and the
  SDK discarded: `Cores` and `RAM` list the values
  `UpgradeComponents` validates against, and `Plan` lists the product
  codes this server can be moved to with `Upgrade`. Without them a
  caller cannot choose a legal value, which is why a cores upgrade
  looked categorically impossible when it was simply outside the
  allowed set for that plan. Also documented that only resizable disks
  appear in `Disk` — a swap partition is absent rather than present and
  empty.
- `server.Create` honours `ParamsOptions` instead of discarding it. It
  previously hardcoded `params[ipv4]=auto` and silently dropped `IPv4`,
  `IPv6`, `Name`, `ContactID`, `Backup` and `SendEmail`, so two of the
  three IP-allocation paths its own documentation described were
  unreachable. Array fields now use the bracket form the API expects.
  An empty `Params.IPv4` still requests automatic allocation, so
  existing callers are unaffected.
- `UpgradeComponentsResponse.Return.Disk` is `map[string]bool`, not
  `bool`. The API answers per disk label (`{"disk":{"scsi0":true}}`), so
  every disk upgrade previously failed to decode with "cannot unmarshal
  object into Go struct field .return.disk of type bool" — disk
  upgrades did not work through this SDK at all. The test fixture that
  enshrined the wrong shape is corrected too.

### Changed

- **Breaking:** `server.LocationSYD1` is renamed `server.LocationLINSYD1`
  so the two Sydney constants follow one scheme.
- `UpgradeComponentsResponse.Return.Disk` is `shtypes.MaybeBoolMap`,
  which accepts the observed object form and a bare bool. Declaring it
  as a map alone traded one decode failure for another pointing the
  other way, hiding the API's real message behind a JSON type error.
- `server.UpgradeComponents` documents cores and RAM as constrained
  **per server** rather than per product family: read the allowed sets
  from `ListUpgrades`. A plan with no headroom returns only the current
  value, which is why an LHPVS1 looks as though it refuses cores
  outright.
- Doc comments no longer claim product codes are undiscoverable, which
  `ListProducts` made false, and no longer use "standard performance"
  for observations made on legacy Xen (LINVPS). The standard-performance
  (SVS) tier was not tested and is now said to be so.

- **Breaking:** `server.ListImages` now takes a `ListImagesOptions`
  argument. Pass the zero value for the previous behaviour.
- **Breaking:** `UpgradeComponentsResponse.Return.Disk` changed type,
  as above.
- `server.UpgradeComponents` documents what each product family
  accepts. An earlier note claimed high-performance and CCS products
  reject component upgrades outright, generalising from a cores/RAM
  test; disk upgrades work fine on high-performance. Disk growth is
  applied online and immediately there, while standard-performance
  stages it as the partition's `NewSize` for `CommitDiskChanges` to
  apply.
- Package documentation for `server`, `server/firewall` and
  `server/firewall/securitygroups` records what a caller has to know up
  front: that product codes are not discoverable and `CanProvision`
  distinguishes "not offered here" from "offered but full"; that a
  server's name is not the label it was given; that SSH keys must be
  registered *and* passed as content at provision time; the per-second
  rate limit and that it surfaces as HTTP 500; that a server cannot be
  deleted while provisioning; and that the firewall endpoints exist
  only for high-performance products.

## [v0.7.1] - 2026-08-20

### Added

- `shtypes.IsEmptyMapShape` reports whether a raw JSON payload is one of
  the shapes PHP produces for "no rows": absent, `null`, or `[]`. The
  previously private `pkg/api/server` helper now delegates to it.
  
### Fixed

- `models.ErrorResponse.Error()` and transport errors no longer include
  the caller's API key; the key is replaced with `REDACTED` while every
  other query parameter is preserved verbatim (order and duplicates
  included). `models.RedactURL` is exported for callers building their
  own error text. Matching is case-insensitive. A response with no
  request now reports its status code instead of dropping it.
- `shtypes.MaybeString` now decodes through `encoding/json`: string
  escapes are handled, `null` and PHP's `[]`-for-empty decode to `""`,
  and non-scalar values (`bool`, populated arrays, objects) error with a
  type description rather than being silently stringified. Adds
  `String()` and `Int()` accessors.
  
## [v0.7.0] - 2026-07-24
### Added
- Support for SRS domain registry endpoints (lookup, lifecycle, contacts, nameservers, WHOIS, transfers, renewals).
- Support for mail endpoints (domains, aliases, accounts, forwards) and a combined list view.
- Support for cloud stacks, databases, SSH users, server config, images, volumes, and Let's Encrypt SSL.
- Support for server snapshots, IP allocation, and lifecycle endpoints.
- Support for DNS templates, reverse DNS, and SOA endpoints.
- Support for SSL certificate endpoints.
- Support for bandwidth usage endpoints.
- Support for the redirect listing endpoint.
- Support for sub-account discovery for resellers.
- `NewClientWithDiscovery` helper on the info client.

### Fixed
- Canonicalise mangled IPv6 addresses returned by the API.
- Bandwidth response decoding bugs.
- Corrected endpoint paths referenced in image package doc comments.

### Changed
- `GetEmailTemplate` now returns a sentinel error when unsupported.

## [v0.6.0] - 2025-06-17
### Fixed
- Made the core Client prepend client and api key to url, avoiding a resort of parameters.
- Update Go version to 1.24.4
- Update golangci-lint version to v2.1.6
- Update pr make file to use golangci-lint GitHub action.
- Split url helpers and type helpers in to their own packages.

## [v0.5.0] - 2025-06-12
### Added
- Added support for all endpoints under `/server/firewall/`.
- Added support for all endpoints under `/server/firewall/security_groups/`.

### Changed
- Moved from v1.3 of the SiteHost API to v1.5.
- Updated response objects to reflect changes in SiteHost API v1.5, which now returns job references including both the `id` and `type` fields.

## [v0.4.0] - 2024-07-17
### Added
- Added support for all endpoints under `/cloud/db/`.
- Added support for all endpoints under `/cloud/db/grant/`.
- Added support for all endpoints under `/cloud/db/user/`.
- Added support for all endpoints under `/cloud/ssh/user/`.
- Added support for the `/bandwith/get_ip_list.json` endpoint.
- Added support for the `/cloud/stack/image/list_all.json` endpoint.

### Fixed
- Corrected the SSH key update parameters.

### Changed
- Updated Go from 1.19 to 1.22.
- Updated dependencies.
- Moved from v1.2 of the SiteHost API to v1.3.
- Changed the type of the `CustomImageAccess` API request struct fields from `string` to `bool`.

## [v0.3.4] - 2024-03-12
### Added
- Add ability to update SSH Keys.

## [v0.3.3] - 2024-03-12
### Added
- Add ability to create, get and delete SSH Keys.

## [v0.3.2] - 2023-03-22
### Fixed
- Fix a crash when unmarshalling when the `/server` returns a different type for the server disk size depending on the type of server.
- Fix GetRecordWithRecord with default priority.
- Fix ListRecords to remove the final dot in the content value.
- Fix the priority value in the UpdateRecord function.

### Added
- Add some helpers for filtering the DNS list and getting new records since there is no get record end point, and the add api does not add the new record id.
- Add helper for handling boolish results.
- Add image list and get endpoints.

## [v0.3.1] - 2023-02-28
### Added
- Added `/cloud/server`, `/cloud/stack`, `/cloud/stack/environment` endpoints.
- Added `/dns` endpoint.
- Added `/ssh` endpoint.
- Added `/api/get_info` endpoint.

## [v0.3.0] - 2022-12-08
### Added
- GitHub PR actions to run go vet, go lint and go mod tidy.

### Updated
- Updated project layout and structure.
- Moved and refactored server and job api endpoints.
- Updated our code to conform with Golang linter.
- Updated our README to link to our Golang style and our license.
- Upgraded golang v1.19.3.
- Upgraded golangci-lint v1.50.1.

## [v0.2.2] - 2022-05-20
### Added
- Added support for SSH keys when provisioning a server.

## [v0.2.1] - 2022-05-16
### Added
- Added label update function.
