// Package token wraps the SiteHost `/ai/token` API endpoints: issuing,
// reading, reconfiguring and revoking the API tokens that authenticate
// against the SiteHost AI platform (ai.sitehost.nz, OpenAI-compatible).
//
// A token is created with a set of capabilities, each of which unlocks
// one family of AI endpoints, and an optional IP allowlist. Both are
// enforced by the AI platform itself, not merely recorded here:
// examples/ai checks each one by calling ai.sitehost.nz with the token.
//
// The token secret is returned exactly once, by [Client.Add]. Nothing
// else in this API returns it again, so a caller that drops it has to
// issue a new token.
package token
