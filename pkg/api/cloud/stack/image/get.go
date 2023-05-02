package image

import (
	"context"
	"errors"

	"github.com/sitehostnz/gosh/pkg/models"
)

// GetImageByCode returns a image by code.
func (s *Client) GetImageByCode(ctx context.Context, request GetRequest) (response models.Image, err error) {
	images, err := s.ListImages(ctx)
	if err != nil {
		return response, err
	}

	for _, image := range images.Return {
		if image.Code == request.Code {
			return image, nil
		}
	}

	return response, errors.New("image not found")
}
