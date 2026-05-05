// Program bandwidth is a read-only validation of the bandwidth
// package's four read endpoints, with particular attention to the two
// decode-bug shapes that motivate the package's custom UnmarshalJSON
// methods and Number type.
//
// Walks all four endpoints and prints a one-line summary of each.
// Designed to be run against any account — including empty / fresh
// ones — and exit zero. Pre-fix, the response decoders failed on
// empty accounts (return: []) and on any quota row with zero usage
// (used_units returned as JSON number rather than string).
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/bandwidth"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("bandwidth: %v", err)
	}
}

func run() error {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required")
	}
	c, err := api.New(apiKey, clientID)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}
	b := bandwidth.New(c)
	ctx := context.Background()

	// 1. ListIPAddresses — exercises the empty-array decode path on
	// accounts with no allocated IPs (return: []).
	ips, err := b.ListIPAddresses(ctx)
	if err != nil {
		return fmt.Errorf("ListIPAddresses: %w", err)
	}
	log.Printf("✓ ListIPAddresses: %d IP(s)", len(ips.Return))

	// 2. GetUsageSummary — exercises the empty-array decode path on
	// accounts with no bandwidth history.
	sum, err := b.GetUsageSummary(ctx)
	if err != nil {
		return fmt.Errorf("GetUsageSummary: %w", err)
	}
	periods := 0
	for _, byPeriod := range sum.Return {
		periods += len(byPeriod)
	}
	log.Printf("✓ GetUsageSummary: %d IP(s), %d period(s) total", len(sum.Return), periods)

	// 3. ListResources — exercises the mixed-type decode path. Any
	// quota row with zero usage returns used_units as a JSON number
	// rather than the string used for non-zero rows; the Number type
	// accepts both.
	res, err := b.ListResources(ctx)
	if err != nil {
		return fmt.Errorf("ListResources: %w", err)
	}
	groups := len(res.Return)
	totalQuotas, zeroQuotas, nonZeroQuotas := 0, 0, 0
	for _, g := range res.Return {
		for _, q := range g.Quotas {
			totalQuotas++
			if q.UsedUnits == 0 {
				zeroQuotas++
			} else {
				nonZeroQuotas++
			}
		}
	}
	log.Printf("✓ ListResources: %d group(s), %d quota(s) total (%d non-zero, %d zero — both decode shapes)",
		groups, totalQuotas, nonZeroQuotas, zeroQuotas)

	// 4. GetUsageByMonth — sanity check on the per-IP populated-shape
	// decode for one of the windowed endpoints. Skipped when the
	// account has no IPs (the empty-array case is already covered by
	// GetUsageSummary above).
	// Pick an IPv4 entry. IPv6 strings come back from
	// /bandwidth/get_ip_list.json with dots in place of colons
	// (e.g. "2403.7000.8000.300..ce/128") and the API will then
	// reject them as invalid on the way back into get_usage_by_*.
	// Separate quirk; not in scope for this PR.
	var anyIP string
	for k, info := range ips.Return {
		if info.Family == "4" {
			anyIP = k
			break
		}
	}
	if anyIP == "" {
		log.Printf("- GetUsageByMonth: skipped (no IPv4 IPs)")
		return nil
	}
	month, err := b.GetUsageByMonth(ctx, bandwidth.UsageOptions{IPAddr: anyIP})
	if err != nil {
		return fmt.Errorf("GetUsageByMonth: %w", err)
	}
	rows := 0
	for _, byPeriod := range month.Return {
		for _, byClass := range byPeriod {
			rows += len(byClass)
		}
	}
	log.Printf("✓ GetUsageByMonth(%s): %d IP(s), %d (period × class) row(s)", anyIP, len(month.Return), rows)

	return nil
}
