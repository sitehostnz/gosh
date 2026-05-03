package snapshot

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// Snapshot describes a single server snapshot.
	//
	// Mixed value types reflect the API's actual response shape:
	// most numeric / boolean fields are returned as strings ("0",
	// "1") while Size is a JSON number and IsMissing is a real
	// JSON bool. Pending is "1" while a job is in flight and "0"
	// when stable.
	Snapshot struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Device        string `json:"device"`
		MountPoint    string `json:"mountpoint"`
		Size          int64  `json:"size"`
		FSType        string `json:"fstype"`
		DRBD          string `json:"drbd"`
		Parent        string `json:"parent"`
		Pending       string `json:"pending"`
		Created       string `json:"created"`
		Backup        string `json:"backup"`
		DiskTotal     string `json:"disk_total"`
		DiskUsed      string `json:"disk_used"`
		InodesTotal   string `json:"inodes_total"`
		InodesUsed    string `json:"inodes_used"`
		StatsUpdated  string `json:"stats_updated"`
		DiskWarn      string `json:"disk_warn"`
		IsMissing     bool   `json:"is_missing"`
		Expires       string `json:"expires"`
	}

	// ListResponse represents the response from list_all.
	ListResponse struct {
		Return []Snapshot `json:"return"`
		models.APIResponse
	}

	// JobResponse is the shared response shape for write
	// operations (create, delete, restore, set_lifetime). Each
	// queues an asynchronous scheduler job.
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
