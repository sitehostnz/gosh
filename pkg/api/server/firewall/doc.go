// Package firewall represents our SiteHost `/server/firewall` API
// endpoint.
//
// # Not available on every product
//
// Verified live (August 2026): high-performance (HPVS) servers have
// security groups, and legacy Xen (LINVPS) servers do not — Get rejects
// those outright:
//
//	Error: Firewall functionality is not available for this server type.
//
// The standard-performance (SVS) tier was **not tested**, so do not
// assume it behaves like either.
//
// Treat that response as a property of the product rather than a
// failure: a caller iterating over a mixed fleet should skip the server
// and carry on, not abort. server.Get reports ProductType, so the
// distinction can be made before calling if preferred.
//
// # Why this matters before you change a server's addresses
//
// The rescue environment used to repair a guest's network configuration
// is reached over the network, so a security group that drops inbound
// SSH would close that recovery path and leave console access as the
// only way in — discovered at exactly the wrong moment. Reported by
// SiteHost operations rather than verified here, so treat it as a
// caution rather than a tested behaviour; either way, checking the
// groups attached to a server before an address change costs nothing.
//
// Get returns the groups attached to a server, not their rules. Read
// the rules through the securitygroups subpackage.
package firewall
