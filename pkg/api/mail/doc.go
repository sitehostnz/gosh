// Package mail represents our SiteHost `/mail` API endpoint —
// shared mail service inspection (server info, domains, accounts,
// aliases, forwarders).
//
// Every operation is scoped to a specific mail server (the
// server_name parameter, e.g. "sth-mail-air" for the Shared Mail
// Service). Consumers must specify which mail service to operate
// against on every call. For ergonomics where a consumer only
// uses one mail service, NewForServer captures the server_name
// once; calls can omit it from per-request options or override
// per call.
//
// Discoverability: at the time of writing, the SiteHost public
// API does not expose an endpoint for client-side enumeration of
// mapped mail services — consumers must obtain the server_name
// out of band (e.g. from the operator).
package mail
