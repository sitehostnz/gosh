# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `server.ListProducts` wraps `server/products.json`, which lists every
  product orderable at a location with its cores, RAM, disk, bandwidth
  and partitions. The endpoint is undocumented rather than absent; it
  removes the need to hardcode a product code, and its partition list
  names the disk labels `UpgradeComponents` requires, knowable before a
  server exists.
- `server.ListImagesOptions` gives `server.ListImages` the filters the
  API already supported: `Type`, `Location`, `IncludeDisabled`,
  `PageSize`/`PageNumber` and `SortBy`/`SortDir`. This makes the
  high-performance (HPVS) image catalogue reachable for the first
  time — it sits behind `Type: server.ImageTypeHPVMDistro` plus a
  mandatory `Location`. HPVS image codes carry a build date
  (`ubuntu-2404-20260727`), so they must be discovered, not hardcoded.
- `server.ImageTypeDistro`, `ImageTypeHPVMDistro`, `ImageTypeContainer`
  and `ImageTypeApp` name the closed set the type filter accepts.
- `server.ProductTypeHPVS`, `ProductTypeSVS`, `ProductTypeLINVPS` and
  `ProductTypeWINVPS` name the product families
  `models.Server.ProductType` reports. `ProductTypeSVS` matters most:
  it is the family `server.LoginUserFor` deliberately does not cover,
  so a caller needs the constant in order to write that check.
- `server.Location*` constants for the location codes known at the time
  of writing, including `LocationLINSYD1` and `LocationWINSYD1`, which
  share a scheme. `server.ListLocations` remains authoritative.
- `server.LoginUserFor` resolves the SSH account for a product family
  and distro, determined empirically. It declines rather than guesses
  for families that were not tested.
- `server.StatisticTypes` and `server.StatisticParameter` describe what
  `ListStatisticTypes` returns; `GetStatisticsOptions` gained `Type`,
  `Item`, `Start`, `End` and `Compacted`.
- `shtypes.MaybeBoolMap`, with `Accepted` and `AcceptedKey`, decodes a
  per-key flag map that the API may also answer as a bare bool or as
  `[]`.
- `securitygroups.AttachedServer` is the shape the security-group
  endpoints report an attached server as.
- `examples/server` walks the whole server lifecycle as a numbered
  journey and exercises all 36 methods in the namespace. Two of its
  steps check the result somewhere other than the API that was asked
  to do the work: the security-group step opens a TCP connection to
  confirm a rule actually filters, and the snapshot step writes a
  marker file into the guest to confirm a restore actually reverted the
  disk.

### Changed

- **Breaking:** `server.ListImages` now takes a `ListImagesOptions`
  argument. Pass the zero value for the previous behaviour.
- **Breaking:** `UpgradeComponentsResponse.Return.Disk` is
  `shtypes.MaybeBoolMap` rather than `bool`. The API answers per disk
  label, and may answer with a bare bool or `[]` when it has no
  per-label detail to give. Read it through `Accepted` and
  `AcceptedKey` rather than indexing.
- **Breaking:** `ListStatisticTypesResponse.Return` is
  `server.StatisticTypes` rather than `[]string`.
- `models.Container.DockerSize` and `models.StackImageVersion.DockerSize`
  are `shtypes.MaybeBigInt`, because the API sends a bare number on a
  container and a quoted string on a version.
- `models.Container.ImageDetails` is a `json.RawMessage`. Its keys are
  `models.StackImage`'s, but its `labels` is a JSON-encoded string
  where StackImage's is an object, and its `versions` is an object
  where StackImage's is a list, so it cannot be decoded as one.
- `server.UpgradeComponents` documents cores and RAM as constrained
  **per server** rather than per product family: read the allowed sets
  from `ListUpgrades`. A plan with no headroom returns only the current
  value, which is why an LHPVS1 looks as though it refuses cores
  outright.
- Package documentation for `server`, `server/firewall` and
  `server/firewall/securitygroups` records what a caller has to know up
  front: that product codes are discoverable through `ListProducts`
  while `CanProvision` distinguishes "not offered here" from "offered
  but full"; that a server's name is not the label it was given, and
  that nothing renames a server; that SSH keys must be registered *and*
  passed as content at provision time; the per-second rate limit and
  that it surfaces as HTTP 500; that a server cannot be deleted while
  provisioning; and that the firewall endpoints exist only for
  high-performance products.
- Doc comments no longer attribute observations made on legacy Xen
  (LINVPS) to standard performance (SVS). Where the SVS tier was not
  tested, the documentation says so rather than generalising.
- Doc comments across `cloud/db`, `cloud/db/user`, `cloud/ssh/user` and
  `cloud/stack` record observed behaviour: list filters are optional
  but validated, so an empty page never means a filter was ignored, and
  `cloud/stack/get.json` takes `server` rather than `server_name`.
- `dns.GetZone` documents that it is a search, so a name matching
  nothing returns `status:true` with an empty list rather than an
  error. Checking `err` alone reports a zone as present when it is not.
- `dns/template.List` documents that the listing includes SiteHost's
  shared templates alongside the account's, so `DomainCount` on a
  shared row is not an account figure.

### Fixed

- `server.Create` ignored `ParamsOptions` entirely, so the IP
  allocation, backup, contact and SSH-key paths its own documentation
  described were unreachable.
- `UpgradeComponentsResponse.Return.Disk` was declared `bool` while the
  API answers per disk label, so every disk upgrade failed to decode
  and disk upgrades did not work through this SDK at all. The test
  fixture that enshrined the wrong shape is corrected too.
- `server.ListUpgrades` never decoded. `QuotaUsage` fields were `int`
  where the API sends fractional values, and `Return.Cores` was
  `[]int` where the API sends quoted integers. Note `ram` arrives as
  bare numbers in the same object.
- `server.ListStatisticTypes` had never returned a metric name.
  `Return` was `[]string`; the API answers with an object keyed by
  metric name and answers `[]` only on a server with no metrics, so it
  decoded exactly the empty case. The existing test asserted a flat
  list of names the API never sends.
- `server.GetStatistics` could not be called successfully at all.
  `type` is required and the options struct had no field for it. `Item`
  — which partition or interface to report on — travels as
  `options[item]`; sending a `partition` or `iface` parameter instead
  is refused with a message that does not point at the real problem.
- `securitygroups.List` had never decoded. `servers` was declared
  `[]string` while the API sends objects carrying a name and a label.
- `server.ProductAttributes` retained a typed field in `Extra` when the
  API spelled the key with different case, since `encoding/json`
  matches case-insensitively but the cleanup did not.
- `cloud/db.Get` and `cloud/db/user.Get` each sent `client_id` twice
  and added an `api_key` parameter this API does not have, which
  `net.Encode` then dropped for not being in the keys list.
- `cloud/ssh/user.Update` silently ignored `ReadOnlyConfig`. The value
  was added as `params[read_only_config]` while the keys list named
  `params[read_only_config][]`, and `net.Encode` emits only the keys it
  is given — so the field was dropped and the call still succeeded.
- `cloud/db.Add` and `cloud/db.Delete` sent `database` twice.
- `api.Client.NewRequest` set `Content-Type` twice, the first being
  overwritten. Every body this SDK sends is form-encoded.
- `models.CloudServer`, `models.Container`, `models.StackImage`,
  `models.StackImageVersion` and `dns/template.TemplateDetails` now
  decode fields the API sends that no field received. They were being
  dropped in silence.
- `examples/server`: the `delete` step deleted servers it did not
  create, falling back to `SH_SERVER_A`/`SH_SERVER_B` when nothing had
  been provisioned. Deleting a server this process did not create now
  requires `SH_DELETE_SERVERS`.
- `examples/server`: `SH_BASE_URL` was documented in two places and
  read nowhere, so pointing the journey at a sandbox silently ran it
  against production.
- `examples/server`: `SH_SSH_KEY_FILE` was unreachable — the steps that
  need SSH rejected before consulting it.
- `examples/server`: a failure mid-swap left a server holding no
  address with no rollback. Released addresses are now restored on the
  way out, best-effort and loudly.
- `api.Client` retries requests the API rejects for exceeding its
  per-second rate limit. The limit is signalled with HTTP 500 rather
  than 429, so it is indistinguishable from a server error by status
  code — a client that treats it as a failed operation can report a
  build as failed when it never started, or retry a create and make two.
  Retrying is safe even for writes: the limit is applied after the key
  is authenticated but before the request is dispatched, so a throttled
  call never reaches the handler.
- `api.RateLimitError` reports that every attempt was throttled, and
  wraps the last API error so existing `*models.ErrorResponse`
  inspection keeps working through `errors.As`.
- `api.IsRateLimited` tests for a rate-limit rejection, whether retries
  were exhausted or switched off.
- `api.SetRateLimitRetries` and `api.SetRateLimitBackoff` configure the
  behaviour; the defaults are 4 attempts and a 250ms initial backoff,
  doubling to a 1s cap. Retrying honours context cancellation, so a
  caller that gives up is not held by a pending backoff.
- The `api` package documentation now records the limit: it applies per
  reseller rather than per key, defaults to 10 requests per second, and
  is configurable — so clients should not hardcode the default.

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
