package models

type (

	// Job represents reference to a job.
	Job struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	}

	// JobDetails represents the job information.
	JobDetails struct {
		State     string `json:"state"`
		Created   string `json:"created"`
		Started   string `json:"started"`
		Completed string `json:"completed"`
		Logs      []Log  `json:"logs"`
	}

	// Log represents the logs attached to the JobDetails.
	Log struct {
		Date    string `json:"date"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}
)
