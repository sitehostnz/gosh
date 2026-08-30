package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/models"
)

// diskGrowthGB is how much to add when staging a disk resize. Small on
// purpose: the point is to demonstrate the stage/apply pair, not to
// resize anything meaningfully.
const diskGrowthGB = 10

// stepUpgradeDisk stages a disk resize and then applies it, which is
// two calls rather than one.
//
// # Stage, then commit
//
// On legacy Xen (LINVPS), UpgradeComponents does not resize anything.
// It records the intended size, which shows up on the server as
// Partition.NewSize alongside the unchanged Partition.Size, and
// CommitDiskChanges is what actually applies it. High performance
// differs — the resize is online and immediate, with nothing to
// commit; see "On high performance the resize is online" below. This step asserts both halves separately, because a caller that
// only makes the first call sees a successful response and no resize,
// which is a confusing place to end up.
//
// # Never hardcode the disk label
//
// UpgradeComponents.Disk is keyed by the disk's label as the platform
// sees it, and that differs by product and attachment type — "scsi0" on
// high-performance, "xvda1" on Xen, and other values elsewhere. The
// label has to be read from the server's Partitions rather than
// guessed; the API rejects a positional index with "Please specify a
// valid disk label."
//
// # On high performance the resize is online
//
// No reboot is needed: the disk grows while the server runs, verified
// live. SiteHost operations note the platform may need to migrate the
// server to another node to find space, in which case it takes
// considerably longer — not observed here, so poll rather than assuming
// a resize is always quick.
//
// Not part of the default journey: it changes the server's disk and is
// independent of the address swap. Run it on its own.
func stepUpgradeDisk(ctx context.Context, c clients, st *state) error {
	if err := st.resolveServers(); err != nil {
		return err
	}
	name := st.nameA

	target, currentGB, err := findDisk(ctx, c, name)
	if err != nil {
		return err
	}
	wantGB := currentGB + diskGrowthGB
	log.Printf("  %s: disk %q is %dGB, staging %dGB", name, target, currentGB, wantGB)

	applied, err := requestResize(ctx, c, name, target, wantGB)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	// Staged rather than applied, so commit it.
	return commitDisk(ctx, c, name, target, wantGB)
}

// findDisk reads the server's partitions and decides which one to grow,
// returning the label the upgrade endpoint expects.
//
// A server can have more than one data disk, so every partition found is
// logged and the choice is stated rather than made silently. Set
// SH_DISK_LABEL to target a specific one; otherwise the largest
// non-swap partition is used, which is the root disk on a
// single-disk server and may well be the wrong answer on a server with
// additional disks attached.
//
// UpgradeComponents.Disk is a map for this reason: several disks can be
// resized in one call. This step grows one, to keep the stage/apply
// behaviour easy to follow.
func findDisk(ctx context.Context, c clients, name string) (string, int, error) {
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", 0, fmt.Errorf("Get(%s): %w", name, err)
	}
	parts := got.Server.Partitions
	if len(parts) == 0 {
		return "", 0, fmt.Errorf("Get(%s): no partitions reported, so there is no disk label to target", name)
	}

	log.Printf("  %s has %d partition(s):", name, len(parts))
	for _, p := range parts {
		log.Printf("     %-10s size=%-6s new_size=%-6s mountpoint=%q fstype=%s", p.Name, p.Size, p.NewSize, p.Mountpoint, p.Fstype)
	}

	if want := strings.TrimSpace(os.Getenv("SH_DISK_LABEL")); want != "" {
		for _, p := range parts {
			if p.Name != want {
				continue
			}
			n, convErr := strconv.Atoi(strings.TrimSpace(p.Size))
			if convErr != nil {
				return "", 0, fmt.Errorf("partition %q reports an unparseable size %q", want, p.Size)
			}
			log.Printf("  targeting %q (SH_DISK_LABEL)", want)
			return want, n, nil
		}
		return "", 0, fmt.Errorf("SH_DISK_LABEL=%q is not a partition on %s", want, name)
	}

	label, sizeGB, err := pickDisk(parts)
	if err != nil {
		return "", 0, err
	}
	if len(parts) > 2 {
		log.Printf("  targeting %q — the largest non-swap partition. This server has additional disks;", label)
		log.Printf("    set SH_DISK_LABEL if that is not the one you meant")
	} else {
		log.Printf("  targeting %q", label)
	}
	return label, sizeGB, nil
}

// requestResize asks for the new size and returns whether the platform
// applied it immediately or merely staged it.
//
// Both happen, depending on the product:
//
//   - High performance applies it online and immediately. Size reflects
//     the new value straight away, NewSize stays zero, and there is no
//     job to poll.
//   - Standard performance stages it as NewSize, and CommitDiskChanges
//     applies it.
//
// So this reads the partition back rather than assuming either. A
// caller that only ever calls UpgradeComponents works on one platform
// and silently does nothing on the other.
func requestResize(ctx context.Context, c clients, name, target string, wantGB int) (applied bool, err error) {
	time.Sleep(throttle)
	up, err := c.server.UpgradeComponents(ctx, server.UpgradeComponentsRequest{
		Name: name,
		Disk: map[string]int{target: wantGB},
	})
	if err != nil {
		return false, fmt.Errorf("UpgradeComponents(%s): %w", name, err)
	}
	// Disk is keyed by label: the endpoint answers per disk.
	if !up.Return.Disk[target] {
		return false, fmt.Errorf("UpgradeComponents(%s): disk %q was not accepted (disk=%v cores=%v ram=%v)",
			name, target, up.Return.Disk, up.Return.Cores, up.Return.RAM)
	}
	// A job is only present on platforms that do this out of band.
	if err := waitJob(ctx, c.job, up.Return.Job); err != nil {
		return false, fmt.Errorf("UpgradeComponents(%s): %w", name, err)
	}

	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return false, fmt.Errorf("Get(%s): %w", name, err)
	}
	liveGB, _ := strconv.Atoi(strings.TrimSpace(partitionSize(got.Server.Partitions, target)))
	staged := partitionNewSize(got.Server.Partitions, target)

	switch {
	case liveGB >= wantGB:
		log.Printf("✓ applied immediately: disk %q is %dGB, online, no reboot and no commit needed", target, liveGB)
		return true, nil
	case staged != "":
		log.Printf("✓ staged: disk %q NewSize=%s while Size is still %dGB — nothing has resized yet",
			target, staged, liveGB)
		return false, nil
	default:
		return false, fmt.Errorf("%s: disk %q is neither resized (%dGB, wanted %dGB) nor staged; the request was accepted but had no effect",
			name, target, liveGB, wantGB)
	}
}

// commitDisk applies a staged resize and asserts the live size grew.
// Only needed on platforms that stage; see requestResize.
func commitDisk(ctx context.Context, c clients, name, target string, wantGB int) error {
	time.Sleep(throttle)
	commit, err := c.server.CommitDiskChanges(ctx, server.CommitDiskChangesRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("CommitDiskChanges(%s): %w", name, err)
	}
	if err := waitJob(ctx, c.job, commit.Return.Job); err != nil {
		return fmt.Errorf("CommitDiskChanges(%s): %w", name, err)
	}

	time.Sleep(throttle)
	after, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", name, err)
	}
	gotGB, _ := strconv.Atoi(strings.TrimSpace(partitionSize(after.Server.Partitions, target)))
	if gotGB < wantGB {
		return fmt.Errorf("%s: disk %q is %dGB after commit, wanted at least %dGB", name, target, gotGB, wantGB)
	}
	log.Printf("✓ committed: disk %q is now %dGB", target, gotGB)
	return nil
}

// stepUpgradePlan moves a server to a different product code.
//
// The endpoint is /server/upgrade_plan.json, which the SDK exposes as
// Upgrade — note the name mismatch with UpgradeComponents, which wraps
// /server/upgrade.json and changes cores, RAM and disk instead.
//
// # The valid targets are discoverable
//
// server.ListUpgrades reports, for this specific server, the product
// codes it can be moved to — its Plan field. That is the authoritative
// answer and it is server-specific; ListProducts answers the different
// question of what a *location* offers. This step reads Plan and
// reports it, so a caller need not guess.
//
// SH_UPGRADE_PLAN still selects which of them to move to, since only
// the caller knows the intent, but it is checked against Plan first.
//
// It is then validated with CanProvision: that distinguishes "not
// offered at this location" from "offered but currently unplaceable",
// and a plan change to a product with no capacity here would otherwise
// fail further in.
//
// A plan change may take longer than a component upgrade: SiteHost
// operations note that the platform can migrate the server to another
// node to satisfy the new plan. That was not observed here — the one
// upgrade tested (LHPVS1 to LHPVS2) returned no job and completed in
// seconds — so treat a slow upgrade as expected rather than broken, and
// poll GetState rather than assuming either outcome.
//
// Not part of the default journey: it changes what the server costs.
func stepUpgradePlan(ctx context.Context, c clients, st *state) error {
	if err := st.resolveServers(); err != nil {
		return err
	}
	name := st.nameA

	// Ask what this server can actually become before asking for
	// anything.
	time.Sleep(throttle)
	ups, err := c.server.ListUpgrades(ctx, server.ListUpgradesOptions{Name: name})
	if err != nil {
		return fmt.Errorf("ListUpgrades(%s): %w", name, err)
	}
	log.Printf("  %s can move to: %s", name, strings.Join(ups.Return.Plan, " "))
	log.Printf("  cores allowed %v, ram allowed %v — a set holding only the current", ups.Return.Cores, ups.Return.RAM)
	log.Printf("    value means no component upgrade is available for this plan")

	plan := strings.TrimSpace(os.Getenv("SH_UPGRADE_PLAN"))
	if plan == "" {
		return fmt.Errorf("set SH_UPGRADE_PLAN to one of the targets above")
	}
	if !contains(ups.Return.Plan, plan) {
		return fmt.Errorf("%s cannot move to %q; ListUpgrades offers: %s", name, plan, strings.Join(ups.Return.Plan, " "))
	}

	if err := checkPlanAvailable(ctx, c, st, plan); err != nil {
		return err
	}

	time.Sleep(throttle)
	before, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", name, err)
	}
	if before.Server.ProductCode == plan {
		return fmt.Errorf("%s is already on %s; nothing to assert", name, plan)
	}
	beforeCores, beforeRAM := int64(before.Server.Cores), before.Server.RAM
	log.Printf("  %s: %s -> %s (cores=%d ram=%s)", name, before.Server.ProductCode, plan, beforeCores, beforeRAM)

	time.Sleep(throttle)
	resp, err := c.server.Upgrade(ctx, server.UpgradeRequest{Name: name, Plan: plan})
	if err != nil {
		return fmt.Errorf("Upgrade(%s, %s): %w", name, plan, err)
	}
	if !resp.Status {
		return fmt.Errorf("Upgrade(%s, %s): %s", name, plan, resp.Msg)
	}

	return confirmPlan(ctx, c, name, plan, beforeCores, beforeRAM)
}

// confirmPlan reads the server back and asserts both the product code
// and the resources changed.
//
// Upgrade returns no job, so there is nothing to poll — the only way to
// know it took is to look. And the product code changing is not proof
// the resources did, so both are checked: a no-op plan change would
// otherwise read as a success.
func confirmPlan(ctx context.Context, c clients, name, plan string, beforeCores int64, beforeRAM string) error {
	time.Sleep(throttle * 2)
	after, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", name, err)
	}
	if after.Server.ProductCode != plan {
		return fmt.Errorf("%s still reports product %s after the upgrade, wanted %s (a migration may still be in progress — check GetState)",
			name, after.Server.ProductCode, plan)
	}
	afterCores, afterRAM := int64(after.Server.Cores), after.Server.RAM
	if afterCores == beforeCores && afterRAM == beforeRAM {
		return fmt.Errorf("%s reports %s but cores (%d) and ram (%s) are unchanged; the plan change had no effect on resources",
			name, plan, afterCores, afterRAM)
	}
	log.Printf("✓ %s is now on %s: cores %d -> %d, ram %s -> %s",
		name, plan, beforeCores, afterCores, beforeRAM, afterRAM)
	log.Printf("  applied without a job; a plan change can instead migrate the server, which takes longer")
	return nil
}

// contains reports whether the list holds v.
func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// checkPlanAvailable confirms the target product exists at this
// location and has capacity, which distinguishes a wrong code from a
// full location before anything is changed.
func checkPlanAvailable(ctx context.Context, c clients, st *state, plan string) error {
	time.Sleep(throttle)
	pre, err := c.server.CanProvision(ctx, server.CanProvisionOptions{
		Product:  plan,
		Location: st.cfg.location,
		Distro:   st.cfg.distro,
	})
	if err != nil {
		return fmt.Errorf("CanProvision(%s at %s): %w", plan, st.cfg.location, err)
	}
	if !pre.Status {
		return fmt.Errorf("CanProvision(%s at %s): %s", plan, st.cfg.location, pre.Msg)
	}
	log.Printf("  %s is offered at %s with capacity", plan, st.cfg.location)
	return nil
}

// pickDisk returns the label and current size of the largest non-swap
// partition. Swap is skipped because growing it is never what was
// meant; on a multi-disk server the largest is a guess, which is why
// findDisk says which one it chose.
func pickDisk(parts []models.Partition) (label string, sizeGB int, err error) {
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p.Mountpoint), "swap") {
			continue
		}
		n, convErr := strconv.Atoi(strings.TrimSpace(p.Size))
		if convErr != nil || p.Name == "" {
			continue
		}
		if n > sizeGB {
			label, sizeGB = p.Name, n
		}
	}
	if label == "" {
		return "", 0, fmt.Errorf("no resizable partition found among %d reported", len(parts))
	}
	return label, sizeGB, nil
}

// partitionSize returns the live size of the named partition.
func partitionSize(parts []models.Partition, label string) string {
	for _, p := range parts {
		if p.Name == label {
			return p.Size
		}
	}
	return ""
}

// partitionNewSize returns the staged size of the named partition, or
// empty if nothing is staged.
//
// A staged resize shows as NewSize alongside an unchanged Size. Zero and
// a value equal to Size both mean "nothing staged" — the field is not
// cleared to empty.
func partitionNewSize(parts []models.Partition, label string) string {
	for _, p := range parts {
		if p.Name != label {
			continue
		}
		v := strings.TrimSpace(p.NewSize)
		if v == "" || v == "0" || v == p.Size {
			return ""
		}
		return v
	}
	return ""
}
