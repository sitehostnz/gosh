package server

type (
	// GetRequest represents request params for get server endpoint.
	GetRequest struct {
		ServerName string `json:"name"`
	}

	// DeleteRequest represents a request to delete a Server.
	DeleteRequest struct {
		Name string `json:"name"`
	}

	// UpgradeRequest represents a request to upgrade a Server's plan
	// (product code). Wraps /server/upgrade_plan.json — note the
	// historical method-name / endpoint-name mismatch.
	UpgradeRequest struct {
		Name string `json:"name"`
		Plan string `json:"plan"`
	}

	// UpgradeComponentsRequest configures a hardware-component
	// upgrade on a server via /server/upgrade.json. At least one of
	// Cores / RAM / Disk should be set; passing zero / empty for all
	// is a no-op.
	//
	// Disk is the upgrade path most consumers actually want — VPS
	// products commonly grow disk independently of their plan, where
	// Cores/RAM are tied to the product code. CCS products reject
	// component upgrades entirely (use the Upgrade method against
	// /server/upgrade_plan.json for CCS resizing instead).
	UpgradeComponentsRequest struct {
		Name  string `json:"name"`
		Cores int    `json:"upgrade[cores],omitempty"`
		// RAM is the new total RAM amount in GB, expressed as a
		// string per the API's expectation (e.g. "8" or "16").
		RAM string `json:"upgrade[ram],omitempty"`
		// Disk maps each disk's label (the device name as the
		// platform sees it) to the new total size in GB. The
		// label varies by hypervisor / disk attachment type — Xen
		// surfaces "xvda1" / "xvdb1", virtio surfaces "vda" /
		// "vdb", SCSI / SATA surface "sda" / "scsi0", etc.
		//
		// **Don't hardcode a label.** The right value differs per
		// server. Discover dynamically by reading server.Get's
		// Partitions field — each Partition has a Name (the label
		// the upgrade endpoint expects).
		//
		// API expectations confirmed by live probing:
		//   - Passing a scalar: rejected with
		//     "Please specify an array of disk upgrades."
		//   - Passing array indexed by 0/1/...: rejected with
		//     "Please specify a valid disk label." (the index has
		//     to be a real device name, not a position number).
		//   - Correct: keyed by the actual disk label, e.g.
		//     `map[string]int{"xvda1": 80}` for Xen,
		//     `map[string]int{"scsi0": 80}` for SCSI.
		//
		// The public docs example shows `upgrade[disk][0]=10` but
		// the description text reveals the real form:
		// `upgrade[disk][xvda1]=10` — the index is a device name.
		Disk map[string]int `json:"upgrade[disk],omitempty"`
	}

	// UpdateRequest represents a request to update a Server.
	UpdateRequest struct {
		Name  string `json:"name"`
		Label string `json:"label"`
	}

	// CommitDiskChangesRequest represents request params for CommitDiskChanges server endpoint.
	CommitDiskChangesRequest struct {
		ServerName string `json:"name"`
	}

	// CreateRequest represents a request to create a Server.
	CreateRequest struct {
		ClientID    string        `json:"client_id"`
		Label       string        `json:"label"`
		Location    string        `json:"location"`
		ProductCode string        `json:"product_code"`
		Image       string        `json:"image"`
		Params      ParamsOptions `json:"params"`
	}

	// ParamsOptions represents the additional parameters in the
	// request to create a Server.
	//
	// # IP allocation paths
	//
	// IPv4 / IPv6 control how the new server gets its address(es).
	// There are three paths consumers should know about:
	//
	//   1. **Auto-allocation (recommended for most cases).** Pass
	//      []string{"auto"} for IPv4 and/or IPv6. The platform
	//      picks a free address from the location's pool and binds
	//      it to the new server. This is the path the public API
	//      docs explicitly recommend ("simply pass the string
	//      'auto' to automatically assign an IPv4 address").
	//
	//   2. **Specific pre-allocated address.** Pass the address(es)
	//      directly, e.g. []string{"203.0.113.10"}. The address
	//      must already be allocated to the calling client_id —
	//      the platform won't transfer pool IPs at provision time
	//      this way. server.ListIPs(location) returns the IPs
	//      currently allocated to the client; if it returns []
	//      that does NOT mean the pool is exhausted, only that
	//      this client has no allocations there. Use
	//      server.ListLocations to read pool-wide capacity
	//      (`available_ipv4`, `available_ipv6`).
	//
	//   3. **Manual allocation by SiteHost staff.** Reseller-style
	//      arrangements may have IPs reserved for specific clients
	//      via SiteHost ops, then accessed via path (2) once
	//      visible in ListIPs. Out of band relative to the API.
	//
	// Don't conflate "ListIPs returned []" with "pool exhausted" —
	// the wrapper's empty result almost always means "use 'auto'
	// instead, or check ListLocations." A previous gosh session
	// wasted ~30 minutes retrying because of this confusion.
	ParamsOptions struct {
		Name string `json:"name,omitempty"`
		// IPv4 — see the IP-allocation paths section above. Most
		// callers want []string{"auto"}.
		IPv4      []string `json:"ipv4"`
		IPv6      []string `json:"ipv6,omitempty"`
		SSHKeys   []string `json:"ssh_keys,omitempty"`
		ContactID string   `json:"contact_id,omitempty"`
		Backup    string   `json:"backup,omitempty"`
		SendEmail string   `json:"send_email,omitempty"`
	}

	// GetStateOptions represents request params for the get_state
	// endpoint.
	GetStateOptions struct {
		Name string `url:"name"`
	}

	// ListUpgradesOptions represents request params for the
	// list_upgrades endpoint.
	ListUpgradesOptions struct {
		Name string `url:"name"`
	}

	// GenerateNetworkConfigOptions represents request params for the
	// generate_network_config endpoint.
	GenerateNetworkConfigOptions struct {
		Name string `url:"name"`
	}

	// AddIPOptions describes an IP address to add to a server.
	// The API uses "param" (not "address") for the IP value.
	AddIPOptions struct {
		Name string `url:"name"`
		IP   string `url:"param"`
	}

	// RemoveIPOptions describes an IP address to remove from a
	// server. The API uses "address" here (distinct from add_ip's
	// "param") — the inconsistency is the API's, not gosh's.
	RemoveIPOptions struct {
		Name string `url:"name"`
		IP   string `url:"address"`
	}

	// SetPrimaryIPOptions describes the new primary IP for a
	// server. Uses "address" like remove_ip.
	SetPrimaryIPOptions struct {
		Name string `url:"name"`
		IP   string `url:"address"`
	}

	// ChangeStateOptions describes a server state transition.
	// Valid State values: "power_on", "power_off", "rescue_on",
	// "rescue_off", "reboot".
	ChangeStateOptions struct {
		Name  string `url:"name"`
		State string `url:"state"`
	}

	// CanProvisionOptions checks resource availability for
	// provisioning a server. Product, Location, and Distro are
	// required; Arch is optional.
	CanProvisionOptions struct {
		Product  string `url:"product"`
		Location string `url:"location"`
		Distro   string `url:"distro"`
		Arch     string `url:"arch,omitempty"`
	}

	// ListIPsOptions identifies the location whose available IPs
	// to list. The Location field maps to the API's "location"
	// parameter (a location code from server.ListLocations).
	ListIPsOptions struct {
		Location string `url:"location"`
	}

	// ListStatisticTypesOptions identifies the server whose metric
	// types to enumerate. The parameter is "server_name" — distinct
	// from siblings that use plain "name".
	ListStatisticTypesOptions struct {
		ServerName string `url:"server_name"`
	}

	// GetStatisticsOptions identifies the server whose metric values
	// to fetch. Like ListStatisticTypesOptions, the parameter is
	// "server_name".
	GetStatisticsOptions struct {
		ServerName string `url:"server_name"`
	}
)
