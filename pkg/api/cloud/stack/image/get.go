package image

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Get fetches a cloud image with the spcified code. calls list and filters. basically a helper.
func (s *Client) Get(ctx context.Context, request GetRequest) (*models.StackImage, error) {
	listResponse, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, image := range listResponse.Return {
		if image.Code == request.Code {
			return &image, nil
		}
	}

	return nil, nil
}
