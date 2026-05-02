package version

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetBuild fetches the build log for a specific custom-image build
// via /cloud/image/version/get_build.json. Use this to surface CI
// failures to the consumer when ListAll reports build_status="failed".
func (s *Client) GetBuild(ctx context.Context, request GetBuildRequest) (response GetBuildResponse, err error) {
	if request.Code == "" {
		return response, fmt.Errorf("cloud.image.version.GetBuild: Code is required")
	}
	if request.BuildID == "" {
		return response, fmt.Errorf("cloud.image.version.GetBuild: BuildID is required")
	}

	keys := []string{"apikey", "client_id", "code", "build_id"}

	req, err := s.client.NewRequest("GET", "cloud/image/version/get_build.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("code", request.Code)
	v.Add("build_id", request.BuildID)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
