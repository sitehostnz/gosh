package image

import (
	"context"
	"errors"

	"github.com/sitehostnz/gosh/pkg/models"
)

// ListImages returns a list of images.
func (s *Client) ListImages(ctx context.Context) (response GetResponse, err error) {
	uri := "cloud/stack/image/list_all.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}

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
