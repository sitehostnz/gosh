package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Image type values accepted by ListImagesOptions.Type.
//
// The endpoint validates this against a closed set and rejects
// anything else with "Please specify a valid image type." Note that
// the set of accepted *filter* values is narrower than the set of
// values the endpoint *returns*: rows come back with a type of
// "salt-container", but filtering by it is rejected. See the
// discoverability note on [Client.ListImages].
const (
	// ImageTypeDistro is the standard-performance (LINVPS) catalogue.
	ImageTypeDistro = "distro"

	// ImageTypeHPVMDistro is the high-performance (HPVS) catalogue.
	// Filtering by this type requires Location to be set as well.
	ImageTypeHPVMDistro = "hpvm-distro"

	// ImageTypeContainer is accepted as a filter but returns no rows
	// at the locations tested.
	ImageTypeContainer = "container"

	// ImageTypeApp is accepted as a filter but returns no rows at the
	// locations tested.
	ImageTypeApp = "app"
)

// ListImages retrieves server images via "server/list_images.json".
//
// # Two catalogues, and why the default one is not enough
//
// The endpoint serves more than one catalogue and the unfiltered call
// returns only the standard-performance (LINVPS) one. Provisioning a
// high-performance server (an HPVS product code such as LHPVS1) with
// any code from that default listing fails with "The image '<code>'
// could not be found", because HPVS images live in a separate,
// location-scoped catalogue.
//
// To reach it, set both Type and Location:
//
//	opt := server.ListImagesOptions{
//		Type:     server.ImageTypeHPVMDistro,
//		Location: "AKLNCT",
//	}
//
// Location is mandatory for that type. Omitting it returns "You must
// provide a location to view the images for the 'hpvm-distro' type."
// The catalogue genuinely differs per location, so a code valid in one
// location may not exist in another.
//
// HPVS image codes carry a build date — ubuntu-2404-20260727,
// debian-13-20260727, almalinux-10-20260727 — so they cannot be
// guessed or hardcoded, and they change as images are rebuilt. Always
// discover them through this call rather than pinning a literal.
//
// # Still not discoverable
//
// Cloud Container Server images remain unreachable through this
// endpoint. Their codes are shaped like ubuntu-cc-<release>-<YYYYMMDD>
// (e.g. ubuntu-cc-2404-20260323); rows of that kind report a type of
// "salt-container", but "salt-container" is rejected as a filter
// value, and the "container" and "app" types return no rows. See
// https://github.com/sitehostnz/gosh/issues/61.
//
// # Other filter quirks, verified live (August 2026)
//
//   - IncludeDisabled widens the default listing from 28 rows to 105.
//   - PageSize is ignored unless PageNumber is also set.
//   - The API's own filters[os] parameter rejects the very values it
//     returns ("linux", "windows"), so it is deliberately not exposed
//     here.
//   - Location is ignored for the default catalogue and mandatory for
//     ImageTypeHPVMDistro.
//   - A handful of rows come back with empty Code, Distro and Type and
//     an OS of "unknown"; callers should skip rows with an empty Code.
func (s *Client) ListImages(ctx context.Context, opt ListImagesOptions) (response ListImagesResponse, err error) {
	if opt.Type == ImageTypeHPVMDistro && opt.Location == "" {
		return response, fmt.Errorf("server.ListImages: Location is required when Type is %q", ImageTypeHPVMDistro)
	}

	u := "server/list_images.json"
	keys := []string{"apikey", "client_id"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	for _, f := range opt.filters() {
		v.Add(f[0], f[1])
		keys = append(keys, f[0])
	}
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}

// filters renders the options as ordered wire parameters. Only fields
// the caller set are emitted, so the zero value produces the default
// catalogue rather than a query full of empty filters.
func (o ListImagesOptions) filters() [][2]string {
	var out [][2]string
	add := func(k, v string) { out = append(out, [2]string{k, v}) }

	if o.Type != "" {
		add("filters[type]", o.Type)
	}
	if o.Location != "" {
		add("filters[location]", o.Location)
	}
	if o.IncludeDisabled {
		add("filters[include_disabled]", "1")
	}
	if o.PageSize > 0 {
		add("filters[page_size]", strconv.Itoa(o.PageSize))
	}
	if o.PageNumber > 0 {
		add("filters[page_number]", strconv.Itoa(o.PageNumber))
	}
	if o.SortBy != "" {
		add("filters[sort_by]", o.SortBy)
	}
	if o.SortDir != "" {
		add("filters[sort_dir]", o.SortDir)
	}
	return out
}
