package version

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListAll lists versions (build history) for a custom image via
// /cloud/image/version/list_all.json. ImageID is required.
//
// Each Version carries a BuildStatus ("success" / "failed" /
// "running") and a BuildID. Only the latest successful build is
// available to deploy in a container.
func (s *Client) ListAll(ctx context.Context, request ListAllRequest) (response ListAllResponse, err error) {
	if request.ImageID == 0 {
		return response, fmt.Errorf("cloud.image.version.ListAll: ImageID is required")
	}

	keys := []string{"apikey", "client_id", "image_id"}

	req, err := s.client.NewRequest("GET", "cloud/image/version/list_all.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("image_id", strconv.Itoa(request.ImageID))
	if request.SortBy != "" {
		v.Add("filters[sort_by]", request.SortBy)
		keys = append(keys, "filters[sort_by]")
	}
	if request.SortDir != "" {
		v.Add("filters[sort_dir]", request.SortDir)
		keys = append(keys, "filters[sort_dir]")
	}
	if request.PageSize != 0 {
		v.Add("filters[page_size]", strconv.Itoa(request.PageSize))
		keys = append(keys, "filters[page_size]")
	}
	if request.PageNumber != 0 {
		v.Add("filters[page_number]", strconv.Itoa(request.PageNumber))
		keys = append(keys, "filters[page_number]")
	}
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
