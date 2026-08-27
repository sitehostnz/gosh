package server

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListProducts returns the products a client can order at a location,
// via "server/products.json".
//
// # This endpoint is undocumented, not absent
//
// Nothing in the public API documentation mentions it — requesting its
// docs page returns "the page you are trying to access does not exist" —
// and it is missing from the endpoint listing. It is nonetheless part of
// the public API: it has a generated sapi class, and the platform
// changelog records it as "added new /server/products/ endpoint to the
// public API which lists products available to a client in the provided
// location".
//
// Being undocumented has a cost worth knowing about: the Knowledge Base
// product-code page is maintained by hand, and every client that needs
// a product code has been hardcoding one. If you were about to do the
// same, use this instead.
//
// # Location is required
//
// Products are scoped to a location's product group, so there is no
// "list everything" call. Omitting Location returns "The location
// filter is missing."
//
// That scoping is also the answer to why a code can appear with
// different specifications in different places: each location's group
// holds its own row. Asked per location, a code has exactly one
// specification, which is what makes this endpoint more trustworthy
// than any table.
//
// # What it returns
//
// Code, type, name, price and description, plus per-product Attributes.
// The attribute set varies by product type — a virtual server reports
// cores, RAM, bandwidth, disk and its partitions, while a disk add-on
// reports only disk, and mail products report mailbox counts.
//
// [Product.Attributes] carries the fields worth typing; anything else is
// available in [ProductAttributes.Extra].
//
// Note the partitions: they name the disk labels ("scsi0" on
// high-performance, "xvda1" on Xen) that [Client.UpgradeComponents]
// requires. Those can therefore be known before a server exists, rather
// than only by reading [Client.Get] afterwards.
func (s *Client) ListProducts(ctx context.Context, opt ListProductsOptions) (response ListProductsResponse, err error) {
	if opt.Location == "" {
		return response, fmt.Errorf("server.ListProducts: Location is required")
	}

	u := "server/products.json"
	keys := []string{"apikey", "client_id", "location"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("location", opt.Location)
	for _, t := range opt.Types {
		v.Add("filters[type][]", t)
		keys = append(keys, "filters[type][]")
	}
	for _, c := range opt.Codes {
		v.Add("filters[code][]", c)
		keys = append(keys, "filters[code][]")
	}
	req.URL.RawQuery = net.Encode(v, dedupeKeys(keys))

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}

// dedupeKeys collapses repeated key names while preserving order.
//
// net.Encode emits every value held against a key each time that key
// appears in the list, so a key repeated per value would duplicate the
// whole set.
func dedupeKeys(keys []string) []string {
	seen := make(map[string]bool, len(keys))
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out
}
