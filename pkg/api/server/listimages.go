package server

import "context"

// ListImages retrieves the list of available server images via
// "server/list_images.json".
//
// **Discoverability gap (verified live, May 2026):** this endpoint
// returns only the older `ubuntu-<release>.amd64.cloud` salt-based
// images (focal, xenial, trusty, etc.). The current Cloud
// Container Server image codes — shaped like
// `ubuntu-cc-<release>-<YYYYMMDD>` (e.g. `ubuntu-cc-2404-20260323`)
// — are **not** returned here.
//
// Provisioning a CCS requires the cc-shaped code; pass it as the
// Image field on server.CreateRequest. The current valid code can
// only be discovered via SiteHost staff scheduler-table lookup or
// empirically (running examples/probe-tls-default with a known
// guess). See docs/open-api-questions.md "CCS image catalogue".
func (s *Client) ListImages(ctx context.Context) (response ListImagesResponse, err error) {
	u := "server/list_images.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
