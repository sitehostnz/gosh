package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/api/server/snapshot"
)

// snapshotLifetime is how long the snapshot this step takes should
// live, in hours. Short, because the step deletes it anyway and a
// leftover should expire on its own if the delete fails.
const snapshotLifetime = 1

// stepSnapshot walks a snapshot through its life: take one, change how
// long it lives, restore the server from it, delete it.
//
// # Restore is destructive, and that is the point of running it here
//
// Restoring reverts the server's disk to the snapshot's contents;
// anything written since is lost. That is precisely why it belongs in a
// journey against a server this process provisioned and is about to
// delete, rather than in a caller's first experiment against something
// they care about.
//
// It refuses to run against a server it did not create, for the same
// reason the delete step does. Naming a server for a read-only step is
// not consent to roll its disk back.
//
// # The partition comes from the product, not from a guess
//
// Snapshots are per disk, and the disk labels differ by platform —
// "scsi0" on high performance, "xvda1" on legacy Xen. They are readable
// from server.ListProducts before a server exists, and from server.Get
// afterwards; this step reads them rather than hardcoding one.
func stepSnapshot(ctx context.Context, c clients, st *state) error {
	if len(st.created) == 0 {
		return fmt.Errorf("the snapshot step restores a server, which reverts its disk; run the journey so it acts on a server this process created")
	}
	name := st.created[0]

	partition, err := firstPartition(ctx, c, name)
	if err != nil {
		return err
	}

	// Put a known file in the guest first, so the restore can be
	// checked against the disk rather than against the job's own
	// account of itself.
	//
	// A key is required rather than optional. Skipping the marker when
	// there is no in-process key would silently reduce this step to
	// the job-status assertion it was written to replace, and report
	// green — and it would do so in exactly the configuration the
	// README recommends for standalone runs, where the key comes from
	// SH_SSH_KEY_FILE. requireKey accepts either source.
	if err := requireKey(st, "snapshot"); err != nil {
		return err
	}
	if err := st.ensureAddresses(ctx, c); err != nil {
		return err
	}
	addr := st.ipA.IPAddr
	marker, err := writeMarker(st, addr)
	if err != nil {
		return err
	}
	log.Printf("  wrote %s inside %s", markerPath, name)

	log.Printf("  taking a snapshot of %s partition %s", name, partition)

	id, err := takeSnapshot(ctx, c, name, partition)
	if err != nil {
		return err
	}

	// Remove it whatever happens next. A snapshot left behind consumes
	// storage until its lifetime expires.
	defer func() {
		time.Sleep(throttle)
		resp, err := c.snap.Delete(ctx, snapshot.SnapshotOptions{Name: name, Snapshot: id})
		if err != nil {
			log.Printf("! could not delete snapshot %s on %s: %v", id, name, err)
			return
		}
		if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
			log.Printf("! snapshot %s delete job did not complete: %v", id, err)
			return
		}
		log.Printf("✓ deleted snapshot %s", id)
	}()

	if err := adjustLifetime(ctx, c, name, id); err != nil {
		return err
	}
	return restoreFrom(ctx, c, name, id, marker)
}

// markerPath is where the proof-of-restore file lives inside the guest.
//
// In the login user's home rather than somewhere under /root or /tmp:
// the login account is not root on every platform — "ubuntu" on
// high-performance, "root" on legacy Xen — and /tmp may be cleared on
// boot, which would make the check fail for a reason unrelated to the
// snapshot.
const markerPath = "$HOME/.gosh-snapshot-marker"

// writeMarker puts a known file inside the guest before the snapshot is
// taken, and returns its contents.
//
// # Why bother
//
// Everything else in this step asks the API whether the API did what
// the API was asked. The restore job reports Completed in about ten
// seconds, which is either impressively fast or not a disk revert at
// all, and no amount of reading the job's own status can tell the two
// apart.
//
// So: write a file, snapshot, delete the file, restore, look for the
// file. If it comes back, the disk really was reverted. If it does not,
// the job completed and nothing happened — which is worth knowing and
// is invisible from the control plane.
func writeMarker(st *state, addr string) (string, error) {
	content := "gosh-" + st.cfg.location
	if _, err := sshRun(addr, fmt.Sprintf("printf %%s %s > %s && sync", content, markerPath)); err != nil {
		return "", fmt.Errorf("writing the marker into the guest: %w", err)
	}
	return content, nil
}

// firstPartition reads a disk label off the server itself.
func firstPartition(ctx context.Context, c clients, name string) (string, error) {
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", fmt.Errorf("server.Get: %w", err)
	}
	if len(got.Server.Partitions) == 0 {
		return "", fmt.Errorf("%s reports no partitions; there is nothing to snapshot", name)
	}
	return got.Server.Partitions[0].Name, nil
}

// takeSnapshot creates one and returns its id.
func takeSnapshot(ctx context.Context, c clients, name, partition string) (string, error) {
	time.Sleep(throttle)
	resp, err := c.snap.Create(ctx, snapshot.CreateOptions{
		Name:      name,
		Partition: partition,
		Lifetime:  snapshotLifetime,
	})
	if err != nil {
		return "", fmt.Errorf("snapshot.Create: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return "", fmt.Errorf("snapshot.Create: %w", err)
	}

	// Create's response carries a job, not the snapshot id, so the
	// listing is how a caller learns what was made.
	time.Sleep(throttle)
	list, err := c.snap.List(ctx, snapshot.ListOptions{Name: name})
	if err != nil {
		return "", fmt.Errorf("snapshot.List: %w", err)
	}
	if len(list.Return) == 0 {
		return "", fmt.Errorf("snapshot.Create reported success but %s has no snapshots", name)
	}
	id := list.Return[len(list.Return)-1].ID
	log.Printf("✓ snapshot %s taken (%d on this server)", id, len(list.Return))
	return id, nil
}

// adjustLifetime changes how long the snapshot lives and reads it back.
func adjustLifetime(ctx context.Context, c clients, name, id string) error {
	const extended = snapshotLifetime + 1

	time.Sleep(throttle)
	resp, err := c.snap.SetLifetime(ctx, snapshot.SetLifetimeOptions{
		Name:     name,
		Snapshot: id,
		Lifetime: extended,
	})
	if err != nil {
		return fmt.Errorf("snapshot.SetLifetime: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("snapshot.SetLifetime: %w", err)
	}
	log.Printf("✓ lifetime set to %dh", extended)
	return nil
}

// restoreFrom rolls the server back to the snapshot, and checks the
// guest's disk actually changed.
func restoreFrom(ctx context.Context, c clients, name, id, marker string) error {
	// Remove the marker so its reappearance means something. Without
	// this the check would pass whether or not the restore did
	// anything, which is the shape of assertion this journey exists to
	// avoid.
	addr, err := guestAddress(ctx, c, name)
	if err != nil {
		return err
	}
	// Report the resolved path, so a failure here names the file it
	// was actually looking at.
	resolved, err := sshRun(addr, "printf %s "+markerPath)
	if err != nil {
		return fmt.Errorf("resolving the marker path: %w", err)
	}
	resolved = strings.TrimSpace(resolved)

	if _, err := sshRun(addr, "rm -f "+markerPath+"; sync"); err != nil {
		return fmt.Errorf("removing the marker before the restore: %w", err)
	}

	// Read the answer from the exit status rather than from stdout.
	// Parsing output for a word meant reading an empty string as "the
	// file is still there", which is a third possible answer the check
	// was not written to expect — and it reported the wrong one of the
	// two it did expect.
	if exists(addr, markerPath) {
		return fmt.Errorf("the marker at %s is still present after deleting it; the check below would prove nothing", resolved)
	}
	log.Printf("  removed the marker; it is absent from the running disk")

	return doRestore(ctx, c, name, id, marker, addr)
}

// doRestore performs the restore and verifies it.
func doRestore(ctx context.Context, c clients, name, id, marker, addr string) error {
	time.Sleep(throttle)
	resp, err := c.snap.Restore(ctx, snapshot.SnapshotOptions{Name: name, Snapshot: id})
	if err != nil {
		return fmt.Errorf("snapshot.Restore: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("snapshot.Restore: %w", err)
	}
	log.Printf("✓ restore job completed for %s from snapshot %s", name, id)

	// The real check: is the file back?
	if err := assertMarkerRestored(addr, marker); err != nil {
		return err
	}

	// The snapshot survives a restore — restoring is not consuming it,
	// which is worth knowing before assuming cleanup is automatic.
	time.Sleep(throttle)
	list, err := c.snap.List(ctx, snapshot.ListOptions{Name: name})
	if err != nil {
		return fmt.Errorf("snapshot.List after restore: %w", err)
	}
	for _, s := range list.Return {
		if s.ID == id {
			log.Printf("  the snapshot still exists after the restore; it is not consumed")
			return nil
		}
	}
	log.Printf("  note: the snapshot is gone after the restore, so a restore consumes it")
	return nil
}

// guestAddress returns the server's current primary address.
func guestAddress(ctx context.Context, c clients, name string) (string, error) {
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", fmt.Errorf("server.Get: %w", err)
	}
	if len(got.Server.Ips) == 0 {
		return "", fmt.Errorf("%s holds no addresses", name)
	}
	return got.Server.Ips[0].IPAddr, nil
}

// assertMarkerRestored checks the file the snapshot captured is back on
// the running disk.
//
// This is the assertion that distinguishes a restore from a job that
// merely reports success. It polls, because the guest may be settling
// after the revert, and it reports how long convergence took rather
// than implying it was instant.
func assertMarkerRestored(addr, want string) error {
	deadline := time.Now().Add(restoreSettle)
	start := time.Now()
	var last string
	for {
		out, err := sshRun(addr, "cat "+markerPath+" 2>/dev/null || true")
		if err == nil {
			last = strings.TrimSpace(out)
			if last == want {
				log.Printf("  ✓ %s is back on the disk with its original contents (%s)",
					markerPath, time.Since(start).Round(time.Second))
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s did not come back after %s (last read %q, want %q) — the restore job reported success but the guest's disk was not reverted",
				markerPath, time.Since(start).Round(time.Second), last, want)
		}
		time.Sleep(5 * time.Second)
	}
}

// restoreSettle bounds how long to wait for a restored disk to show the
// snapshot's contents.
const restoreSettle = 3 * time.Minute

// exists reports whether a path is present in the guest.
//
// It reads the exit status of test(1) rather than any output, so there
// is no third answer to misinterpret.
func exists(addr, path string) bool {
	_, err := sshRun(addr, "test -e "+path)
	return err == nil
}
