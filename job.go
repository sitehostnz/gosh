package gosh

import (
	"context"
	"fmt"
)

type JobsService service

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

type Log struct {
	Date    string `json:"date"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (s *JobsService) Get(ctx context.Context, id string) (*Job, error) {
	u := fmt.Sprintf("job/get.json?job_id=%s&type=daemon", id)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(Job)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
