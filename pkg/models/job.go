package models

type (
	// JobDetails represents the job information.
	JobDetails struct {
		Created   string `json:"created"`
		Started   string `json:"started"`
		Completed string `json:"completed"`
		Message   string `json:"message"`
		State     string `json:"state"`
		Logs      []Log  `json:"logs"`
	}

	// Log represents the logs attached to the JobDetails.
	Log struct {
		Date    string `json:"date"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}
)
