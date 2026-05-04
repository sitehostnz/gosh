// Package info wraps SiteHost's /api/* endpoints. The endpoint
// exposed today is api/get_info.json (Get), which returns the
// authenticated client's client_id, contact_id, and the list of
// modules and roles their key has access to.
//
// # Bootstrap pattern
//
// api/get_info.json returns the authenticated client's client_id.
// Consumers who do not yet know their client_id — for example a
// freshly-issued super-user key being wired into a tool — can use
// NewClientWithDiscovery to obtain a fully-configured *api.Client
// from just the API key.
//
// For sub-account targeting from a super-user key, prefer
// api.New(apiKey, subAccountClientID, opts...) directly:
// discovery resolves to the super-user's own client_id, which is
// not the sub-account's id and would scope subsequent calls to
// the wrong account.
package info
