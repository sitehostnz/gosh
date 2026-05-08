package models

type (

	// Job represents reference to a job.
	Job struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	}

	// JobDetails represents the job information returned by job/get.json.
	//
	// Message is populated when the job has surfaced a top-level error
	// or status message — typically present on state="Failed" and on
	// some completed-but-noisy outcomes. It mirrors the API's
	// `return.message` field and is the right place to read a
	// human-friendly explanation of why a job failed (e.g. "Image not
	// typed", "job already running on this stack"). Distinct from the
	// per-step Logs entries.
	JobDetails struct {
		State     string `json:"state"`
		Created   string `json:"created"`
		Started   string `json:"started"`
		Completed string `json:"completed"`
		Logs      []Log  `json:"logs"`
		Message   string `json:"message"`
	}

	// Log represents the logs attached to the JobDetails.
	Log struct {
		Date    string `json:"date"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}
)
