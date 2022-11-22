package models

// Job represents a SiteHost Job.
type Job struct {
	Return struct {
		Created   string `json:"created"`
		Started   string `json:"started"`
		Completed string `json:"completed"`
		Message   string `json:"message"`
		State     string `json:"state"`
		Logs      []Log  `json:"logs"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// Log represents the logs attached to the Job.
type Log struct {
	Date    string `json:"date"`
	Level   string `json:"level"`
	Message string `json:"message"`
}
