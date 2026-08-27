package server

import (
	"encoding/json"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type (
	// DeleteResponse represents a result of a delete Server call.
	DeleteResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// CommitDiskChangesResponse represents a result of a commit changes Server call.
	CommitDiskChangesResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// UpgradeComponentsResponse represents a result of an
	// /server/upgrade.json call. Cores and RAM are per-component
	// flags; Disk is keyed by disk label.
	//
	// Disk is a map, not a bool. The endpoint answers per disk,
	// because the request is per disk:
	//
	//	{"return": {"disk": {"scsi0": true}}}
	//
	// Verified live against a high-performance server, August 2026. It
	// was previously declared as a bool here, which made every disk
	// upgrade fail to decode with "cannot unmarshal object into Go
	// struct field .return.disk of type bool" — so disk upgrades did
	// not work through this SDK at all.
	//
	// Note also that a response may carry no job: on high-performance
	// products a disk grow is applied online and immediately, so
	// models.Job is zero and there is nothing to poll. See
	// [Client.UpgradeComponents].
	UpgradeComponentsResponse struct {
		Return struct {
			models.Job `json:"job"`
			Cores      bool            `json:"cores"`
			RAM        bool            `json:"ram"`
			Disk       map[string]bool `json:"disk"`
		} `json:"return"`
		models.APIResponse
	}

	// CreateResponse represents a result of the create a Server call.
	CreateResponse struct {
		Return struct {
			models.Job `json:"job"`
			Name       string   `json:"name"`
			Password   string   `json:"password"`
			Ips        []string `json:"ips"`
			ServerID   string   `json:"server_id"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse represents a result of a get Server call.
	GetResponse struct {
		Server models.Server `json:"return"`
		models.APIResponse
	}

	// ListResponse lists all servers.
	ListResponse struct {
		Return struct {
			models.Pagination
			Servers []models.Server `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse represents a result of a update Server call.
	UpdateResponse struct {
		models.APIResponse
	}

	// UpgradeResponse represents a result of a upgrade Server call.
	UpgradeResponse struct {
		models.APIResponse
	}

	// LastJob is a brief summary of the most recent job affecting a
	// server, returned by get_state.
	LastJob struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		State string `json:"state"`
	}

	// ServerState is the runtime state of a server.
	//
	//nolint:revive // name kept verbose for grep-from-API parity
	ServerState struct {
		State   string  `json:"state"`
		Rescue  bool    `json:"rescue"`
		LastJob LastJob `json:"last_job"`
	}

	// GetStateResponse represents the response from get_state.
	GetStateResponse struct {
		Return ServerState `json:"return"`
		models.APIResponse
	}

	// Image is a server image entry from list_images.
	Image struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Arch   string `json:"arch"`
		Distro string `json:"distro"`
		Type   string `json:"type"`
		OS     string `json:"os"`
	}

	// ListImagesResponse represents the response from list_images.
	ListImagesResponse struct {
		Return []Image `json:"return"`
		models.APIResponse
	}

	// Location is a datacenter location entry from list_locations.
	// Public is returned as a string flag ("0"/"1") by the API.
	Location struct {
		Public             string   `json:"public"`
		OS                 []string `json:"os"`
		Label              string   `json:"label"`
		Code               string   `json:"code"`
		Datacenter         string   `json:"datacenter"`
		AvailableIPs       int      `json:"available_ips"`
		AvailableIPv4      int      `json:"available_ipv4"`
		AvailableIPv6      int      `json:"available_ipv6"`
		IPv6               bool     `json:"ipv6"`
		PublicPrivateCloud bool     `json:"public_private_cloud"`
		ProductTypes       []string `json:"product_types"`
	}

	// ListLocationsResponse represents the response from list_locations.
	ListLocationsResponse struct {
		Return []Location `json:"return"`
		models.APIResponse
	}

	// ResourceQuota is a single quota entry inside a resource group.
	// AvailableUnits is returned as a number (and may be negative
	// when over-quota).
	//
	// TotalUnits and UsedUnits use [Number] because the API mixes
	// JSON-string and JSON-number forms within a single response —
	// see the [Number] type documentation. Same quirk as the
	// parallel /bandwidth/list_resources endpoint addressed in #43.
	ResourceQuota struct {
		AttributeID    string   `json:"attribute_id"`
		AttributeName  string   `json:"attribute_name"`
		AttributeUnit  string   `json:"attribute_unit"`
		AttributeType  string   `json:"attribute_type"`
		TotalUnits     Number   `json:"total_units"`
		UsedUnits      Number   `json:"used_units"`
		AvailableUnits int      `json:"available_units"`
		Objects        []string `json:"objects"`
	}

	// ResourceGroup represents a per-client resource quota group from
	// list_resources.
	ResourceGroup struct {
		ClientID  string          `json:"client_id"`
		GroupID   string          `json:"group_id"`
		GroupName string          `json:"group_name"`
		Quotas    []ResourceQuota `json:"quotas"`
	}

	// ListResourcesResponse represents the response from list_resources.
	ListResourcesResponse struct {
		Return []ResourceGroup `json:"return"`
		models.APIResponse
	}

	// QuotaUsage is a total/used pair returned within UpgradeQuota.
	//
	// Both are float64 because the RAM quota is fractional — a live
	// response reported total 67.5 and used 67.5 gigabytes. Declaring
	// these as int made [Client.ListUpgrades] fail to decode entirely
	// ("cannot unmarshal number 67.5 into ... of type int"), so that
	// endpoint had never worked. Corrected August 2026.
	//
	// Used can exceed Total when an account is over quota; that is a
	// reportable state, not a decode problem.
	QuotaUsage struct {
		Total float64 `json:"total"`
		Used  float64 `json:"used"`
	}

	// UpgradeQuota is the overall quota / usage block from list_upgrades.
	UpgradeQuota struct {
		RAM   QuotaUsage `json:"ram"`
		Disk  QuotaUsage `json:"disk"`
		Cores QuotaUsage `json:"cores"`
	}

	// ExtraDiskOption is the per-unit price/size offered for additional
	// disk capacity.
	ExtraDiskOption struct {
		Price float64 `json:"price"`
		Size  int     `json:"size"`
	}

	// DiskUpgradeOptions is the per-disk-slot list of included and
	// available extra disk sizes.
	DiskUpgradeOptions struct {
		Included []int `json:"included"`
		Extra    []int `json:"extra"`
	}

	// Upgrades is the upgrade-availability information for a server,
	// returned by [Client.ListUpgrades].
	//
	// This is the endpoint that says what an upgrade may legally ask
	// for, and [Client.UpgradeComponents] validates against exactly
	// these values. Read it first rather than guessing: a request
	// outside these sets is rejected with "Please specify a valid cores
	// value." (or ram, or disk label, or disk size), which does not say
	// what would have been valid.
	Upgrades struct {
		Quota     UpgradeQuota    `json:"quota"`
		ExtraDisk ExtraDiskOption `json:"extra-disk"`

		// Disk lists the sizes each disk may be grown to, keyed by disk
		// label (e.g. "scsi0").
		//
		// Only resizable disks appear. A server's swap partition is
		// absent from this map rather than present with empty options,
		// so treat a missing label as "not resizable" — do not assume
		// every label from models.Server.Partitions has an entry here.
		Disk map[string]DiskUpgradeOptions `json:"disk"`

		// Cores lists the core counts this server may be set to. It
		// contains only the current value when no change is available,
		// which is why a cores upgrade can be rejected on a server
		// whose platform supports them in general — the constraint is
		// per server, not per product family.
		Cores []int `json:"cores"`

		// RAM lists the memory sizes in gigabytes this server may be set
		// to. Fractional values occur, so this is a float slice. As with
		// Cores, it may contain only the current value.
		RAM []float64 `json:"ram"`

		// Plan lists the product codes this server can be moved to with
		// [Client.Upgrade]. This is the authoritative answer to "what
		// can this server become" — server-specific, and not derivable
		// from ListProducts, which answers what a *location* offers.
		//
		// It can include products that are not generally orderable, so
		// treat it as what the platform will accept rather than as a
		// customer-facing menu.
		Plan []string `json:"plan"`
	}

	// ListUpgradesResponse represents the response from list_upgrades.
	ListUpgradesResponse struct {
		Return Upgrades `json:"return"`
		models.APIResponse
	}

	// GenerateNetworkConfigResponse represents the response from
	// generate_network_config. The Return map is keyed by file path
	// (e.g. "/etc/netplan/50-cloud-init.yaml") with file contents
	// as the value.
	GenerateNetworkConfigResponse struct {
		Return map[string]string `json:"return"`
		models.APIResponse
	}

	// IPJobResponse represents the shared response from add_ip and
	// remove_ip — a scheduler job plus the IP address that was
	// affected.
	IPJobResponse struct {
		Return struct {
			models.Job `json:"job"`
			IPAddr     string `json:"ip_addr"`
		} `json:"return"`
		models.APIResponse
	}

	// SetPrimaryIPResponse represents the synchronous response
	// from set_primary_ip.
	SetPrimaryIPResponse struct {
		Return struct {
			IPAddr string `json:"ip_addr"`
		} `json:"return"`
		models.APIResponse
	}

	// ChangeStateResponse represents the response from
	// change_state — a scheduler job for the state transition.
	ChangeStateResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}

	// AvailableIP is a single IP entry from server.list_ips. The
	// API returns objects, not strings — earlier versions of this
	// wrapper had `Return []string` which silently dropped the
	// per-IP shape.
	AvailableIP struct {
		IPAddr string `json:"ip_addr"`
		Prefix int    `json:"prefix"`
		Family int    `json:"family"` // 4 for IPv4, 6 for IPv6
	}

	// ListIPsResponse is the response for server.list_ips. The
	// Return slice lists IPs available for new provisioning at the
	// given location; empty when none are free.
	ListIPsResponse struct {
		Return []AvailableIP `json:"return"`
		models.APIResponse
	}

	// AllocatedIP describes one IP currently allocated to the
	// authenticated client, as returned by list_allocated_i_ps.
	AllocatedIP struct {
		IPAddr   string `json:"ip_addr"`
		Netmask  string `json:"netmask"`
		Gateway  string `json:"gateway"`
		Location string `json:"location"`
		Type     string `json:"type"` // "v4" or "v6"
	}

	// ListAllocatedIPsResponse is the response for
	// list_allocated_i_ps. Return is a map keyed by a transformed IP
	// string (IPv4 dots preserved; IPv6 colons replaced with dots,
	// double-colon replaced with double-dot). The IPAddr field on
	// each value carries the original IP literal.
	ListAllocatedIPsResponse struct {
		Return map[string]AllocatedIP `json:"return"`
		models.APIResponse
	}

	// ListStatisticTypesResponse is the response for
	// list_statistic_types. Return enumerates the metric type IDs
	// the named server currently exposes.
	ListStatisticTypesResponse struct {
		Return []string `json:"return"`
		models.APIResponse
	}

	// GetStatisticsResponse is the response for get_statistics. The
	// Return shape is server-specific and time-windowed; consumers
	// typically deserialise selectively from the raw JSON when
	// looking at specific metrics. Captured here as a generic map
	// to keep gosh's surface usable without schema-locking.
	GetStatisticsResponse struct {
		Return map[string]interface{} `json:"return"`
		models.APIResponse
	}
)

// isEmptyMapShape reports whether the raw "return" payload is one of
// the forms the API uses for an empty map: omitted/zero-length, JSON
// null, or the empty array `[]` it serialises in place of `{}`.
func isEmptyMapShape(raw json.RawMessage) bool {
	return shtypes.IsEmptyMapShape(raw)
}

// UnmarshalJSON tolerates the empty-array form the API returns when
// the server has no generated network config rows.
func (r *GenerateNetworkConfigResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse
	raw := envelope.Return
	if isEmptyMapShape(raw) {
		r.Return = map[string]string{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}

// UnmarshalJSON tolerates the empty-array form the API returns when
// the server has no allocated IPs.
func (r *ListAllocatedIPsResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse
	raw := envelope.Return
	if isEmptyMapShape(raw) {
		r.Return = map[string]AllocatedIP{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}

// UnmarshalJSON tolerates the empty-array form the API returns when
// no statistics are available for the requested interval.
func (r *GetStatisticsResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse
	raw := envelope.Return
	if isEmptyMapShape(raw) {
		r.Return = map[string]interface{}{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}
