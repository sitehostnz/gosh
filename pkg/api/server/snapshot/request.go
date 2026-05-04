package snapshot

type (
	// ListOptions identifies the server whose snapshots to list.
	// Note the API uses the unprefixed parameter name "name" for
	// the server identifier (not "server_name").
	ListOptions struct {
		Name string `url:"name"`
	}

	// CreateOptions describes a new snapshot to take. All fields
	// are required: Name (the server), Partition (disk slot, e.g.
	// "scsi0"), and Lifetime (in hours).
	CreateOptions struct {
		Name      string `url:"name"`
		Partition string `url:"partition"`
		Lifetime  int    `url:"lifetime"`
	}

	// SnapshotOptions identifies a snapshot for the operations
	// that act on a single snapshot (delete, restore). The server
	// is identified by Name; the snapshot by Snapshot.
	//
	//nolint:revive // name kept verbose for grep-from-API parity
	SnapshotOptions struct {
		Name     string `url:"name"`
		Snapshot string `url:"snapshot"`
	}

	// SetLifetimeOptions describes a lifetime adjustment for an
	// existing snapshot. Lifetime is in hours.
	SetLifetimeOptions struct {
		Name     string `url:"name"`
		Snapshot string `url:"snapshot"`
		Lifetime int    `url:"lifetime"`
	}
)
