package server

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents a server response for listing stack servers.
	ListResponse struct {
		CloudServers []models.CloudServer `json:"return"`
		models.APIResponse
	}

	// UpdateWindow describes a CCS's maintenance/patching window
	// configuration as returned by get_update_window. DayOfWeek
	// is 1–7 (1 = Monday); HourOfDay is 0–23; MinuteOfHour is 0–59.
	UpdateWindow struct {
		Enabled      bool `json:"enabled"`
		DayOfWeek    int  `json:"day_of_week"`
		HourOfDay    int  `json:"hour_of_day"`
		MinuteOfHour int  `json:"minute_of_hour"`
	}

	// GetUpdateWindowResponse is the response from
	// cloud/server/get_update_window.json. Returns the current
	// maintenance-window configuration.
	GetUpdateWindowResponse struct {
		Return UpdateWindow `json:"return"`
		models.APIResponse
	}

	// JobResponse is the shared response shape for the asynchronous
	// write operations: SetUpdateWindow, UpdateMinimumTLSVersion.
	// Each queues a scheduler job; the job ID is returned for
	// tracking.
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
