// Package srs wraps the SiteHost /srs API — domain registration,
// transfer, contacts, and lifecycle ops via the Shared Registry
// System.
//
// # TLD-policy variation (important)
//
// SiteHost's SRS API is a uniform façade over multiple
// upstream-registry policies — but **not all features apply to
// every TLD** the registry serves. The API doesn't surface a
// per-TLD capability matrix; consumers learn about restrictions
// only when an op rejects with a TLD-specific error message.
//
// Confirmed TLD-specific behaviours (May 2026):
//
//   - **.nz domains reject LockDomain / UnlockDomain** with "This
//     domain cannot be locked." The .nz registry uses the UDAI
//     (transfer authorisation code) model rather than EPP-style
//     transfer locks. Use NewUDAI / ValidateUDAI for transfer
//     auth on .nz instead.
//
//   - Other TLD-specific quirks are likely (privacy-protection
//     not supported on some TLDs, contact-update rules differing
//     between registries, etc.) but haven't been exhaustively
//     catalogued. Treat any "this domain cannot ..." error as a
//     TLD-policy signal, not a wrapper bug.
//
// **Consumer guidance:** for each write op, call once and check
// the response. Don't assume a wrapper that works on .com works
// identically on .nz / .au / .uk / etc. The wrapper produces the
// well-formed request; the API enforces TLD policy on top.
//
// **Open question for the API team:** could there be a per-TLD
// capability endpoint so SDK consumers know upfront which ops will
// be rejected?
//
// # Endpoint surface in this package
//
// Reads: ListDomains, GetDomain, ListContacts, GetContact,
// SearchContacts, ListNameServers, ListValidTLDs,
// GetPricingTiers, GetCompanyInfo, Whois, GetDomainPrice,
// DomainAvailable, DomainInsideGracePeriod, CanTransferDomain.
//
// Writes: CreateDomain, CancelDomain, LockDomain, UnlockDomain,
// EnablePrivacyProtection, DisablePrivacyProtection,
// UpdateAutoRenew. (More writes — contacts CRUD, transfer flow,
// UDAI, name servers — coming in Tier 2+.)
package srs
