// Package securitygroups represents our SiteHost
// `/server/firewall/security_groups` API endpoint.
//
// # Not available on every product
//
// Security groups were verified present on high-performance (HPVS)
// servers and absent on legacy Xen (LINVPS); the standard-performance
// (SVS) tier was not tested. See the parent firewall package.
//
// # Rules
//
// Get returns a group's rules split into inbound (Rules.In) and
// outbound (Rules.Out). When deciding whether a port is reachable,
// note that Rule.DestPort is zero when a rule does not constrain the
// port, and Rule.Protocol is empty when it does not constrain the
// protocol — so a zero value means "any", not "port 0". A rule with
// Enabled false is present but inert.
//
// Distinguish "explicitly allowed" from "no rule mentions this port at
// all": both leave a port open in practice, but only the first is a
// deliberate decision, and reporting the second as a pass is
// misleading.
package securitygroups
