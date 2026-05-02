package volume

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// Volume describes a single cloud volume.
	//
	// String-typed boolean fields (IsMissing) reflect the API's
	// actual response shape ("0"/"1"). Pending is non-empty while
	// an operation is in flight (e.g. "adding:29658430"); empty
	// when stable.
	//
	// Containers is a list of container identifiers the volume is
	// currently attached to; empty when not mounted anywhere.
	Volume struct {
		ID           string   `json:"id"`
		ClientID     string   `json:"client_id"`
		ServerID     string   `json:"server_id"`
		Pending      string   `json:"pending"`
		VolumeName   string   `json:"volume_name"`
		IsMissing    string   `json:"is_missing"`
		DateAdded    string   `json:"date_added"`
		DateUpdated  string   `json:"date_updated"`
		ServerName   string   `json:"server_name"`
		ServerLabel  string   `json:"server_label"`
		ServerOwner  bool     `json:"server_owner"`
		Containers   []string `json:"containers"`
	}

	// ListResponse represents the response from list_all.
	ListResponse struct {
		Return struct {
			models.Pagination
			Data []Volume `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse represents the response from get.
	GetResponse struct {
		Return Volume `json:"return"`
		models.APIResponse
	}

	// JobResponse is the shared response shape for the write
	// operations (add, delete, mount, update_mounts). Each
	// queues an asynchronous scheduler job; the job ID is
	// returned for tracking.
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
