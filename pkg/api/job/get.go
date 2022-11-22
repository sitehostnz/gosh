package job

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Get gets the Job with the provided ID.
func (s *JobsService) Get(ctx context.Context, id string) (*models.Job, error) {
	u := fmt.Sprintf("job/get.json?job_id=%s&type=daemon", id)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(models.Job)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
