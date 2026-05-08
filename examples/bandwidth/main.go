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
	"github.com/sitehostnz/gosh/pkg/models"
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

	ips, err := stepListIPAddresses(ctx, b)
	if err != nil {
		return err
	}
	if err := stepGetUsageSummary(ctx, b); err != nil {
		return err
	}
	if err := stepListResources(ctx, b); err != nil {
		return err
	}
	return stepGetUsageByMonth(ctx, b, ips)
}

// stepListIPAddresses exercises the empty-array decode path on
// accounts with no allocated IPs (return: []).
func stepListIPAddresses(ctx context.Context, b *bandwidth.Client) (map[string]models.IPAddress, error) {
	resp, err := b.ListIPAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListIPAddresses: %w", err)
	}
	log.Printf("✓ ListIPAddresses: %d IP(s)", len(resp.Return))
	return resp.Return, nil
}

// stepGetUsageSummary exercises the empty-array decode path on
// accounts with no bandwidth history.
func stepGetUsageSummary(ctx context.Context, b *bandwidth.Client) error {
	resp, err := b.GetUsageSummary(ctx)
	if err != nil {
		return fmt.Errorf("GetUsageSummary: %w", err)
	}
	periods := 0
	for _, byPeriod := range resp.Return {
		periods += len(byPeriod)
	}
	log.Printf("✓ GetUsageSummary: %d IP(s), %d period(s) total", len(resp.Return), periods)
	return nil
}

// stepListResources exercises the mixed-type decode path. Any quota
// row with zero usage returns used_units as a JSON number rather than
// the string used for non-zero rows; the Number type accepts both.
func stepListResources(ctx context.Context, b *bandwidth.Client) error {
	resp, err := b.ListResources(ctx)
	if err != nil {
		return fmt.Errorf("ListResources: %w", err)
	}
	total, zero, nonZero := 0, 0, 0
	for _, g := range resp.Return {
		for _, q := range g.Quotas {
			total++
			if q.UsedUnits == 0 {
				zero++
			} else {
				nonZero++
			}
		}
	}
	log.Printf("✓ ListResources: %d group(s), %d quota(s) total (%d non-zero, %d zero — both decode shapes)",
		len(resp.Return), total, nonZero, zero)
	return nil
}

// stepGetUsageByMonth is a sanity check on the per-IP populated-shape
// decode. Skipped when the account has no IPv4 IPs.
//
// Filters to IPv4 deliberately: /bandwidth/get_ip_list.json returns
// IPv6 strings with ':' replaced by '.' (and '::' as '..'), and the
// API then rejects the mangled form on the way back into
// get_usage_by_*. See models.IPAddress for the full caveat.
func stepGetUsageByMonth(ctx context.Context, b *bandwidth.Client, ips map[string]models.IPAddress) error {
	var ip string
	for k, info := range ips {
		if info.Family == "4" {
			ip = k
			break
		}
	}
	if ip == "" {
		log.Printf("- GetUsageByMonth: skipped (no IPv4 IPs)")
		return nil
	}
	resp, err := b.GetUsageByMonth(ctx, bandwidth.UsageOptions{IPAddr: ip})
	if err != nil {
		return fmt.Errorf("GetUsageByMonth: %w", err)
	}
	rows := 0
	for _, byPeriod := range resp.Return {
		for _, byClass := range byPeriod {
			rows += len(byClass)
		}
	}
	log.Printf("✓ GetUsageByMonth(%s): %d IP(s), %d (period × class) row(s)", ip, len(resp.Return), rows)
	return nil
}
