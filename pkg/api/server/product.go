package server

import (
	"encoding/json"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type (
	// Product is one orderable product at a location, as returned by
	// [Client.ListProducts].
	Product struct {
		// Code is the value to pass as CreateRequest.ProductCode.
		Code string `json:"code"`

		// Type is the product family — "HPVS", "SVS", "LINVPS",
		// "CLDCON", "DISK" and others. See the ProductType* constants
		// for the virtual-server families.
		Type string `json:"type"`

		// Name is the customer-facing name, e.g. "Linux HPVS - 1 Core".
		Name string `json:"name"`

		// Price is the monthly price, carrying no currency — read it as
		// NZD.
		//
		// [shtypes.MaybeString] because the API is inconsistent about
		// quoting numerics: this arrives as a JSON string today, and
		// the sibling partition size arrives as either form within a
		// single response. Use Price.String() or Price.Int().
		Price shtypes.MaybeString `json:"price"`

		// Description is usually empty.
		Description string `json:"description"`

		// Attributes describes what the product provides. Which fields
		// are populated depends on Type.
		Attributes ProductAttributes `json:"attributes"`
	}

	// ProductAttributes is the per-product specification.
	//
	// The attribute set varies by product type: a virtual server reports
	// cores, RAM, bandwidth, disk and partitions; a disk add-on reports
	// only disk; mail products report mailbox counts. Anything without a
	// field here is kept in Extra rather than discarded.
	ProductAttributes struct {
		// Cores is the vCPU count.
		Cores int `json:"cores"`

		// RAM is in gigabytes, and is not always whole — the smallest
		// standard-performance plan is 1.5GB — so it is a float rather
		// than an int.
		RAM float64 `json:"ram"`

		// Disk is the total disk in gigabytes.
		Disk int `json:"disk"`

		// Bandwidth is the monthly allowance in gigabytes. This is the
		// attribute that varies between locations for the same code.
		Bandwidth int `json:"bandwidth"`

		// Partitions names the disks the product ships with. The Name
		// values are the labels UpgradeComponentsRequest.Disk is keyed
		// by, so a disk upgrade can be prepared before the server
		// exists.
		Partitions []ProductPartition `json:"partitions"`

		// Extra holds attributes with no field above — "ha",
		// "containers", "images", "mail_accounts", "mail_storage",
		// "ssd", "hdd" have all been seen. Kept so an unfamiliar
		// product is not silently reduced to zeroes.
		Extra map[string]json.RawMessage `json:"-"`
	}

	// ProductPartition is one disk a product includes.
	ProductPartition struct {
		// Name is the disk label, e.g. "scsi0" or "xvda1".
		Name string `json:"name"`

		// Type is the storage class, e.g. "ssd".
		Type string `json:"type"`

		// Size is in gigabytes.
		//
		// [shtypes.MaybeString] because the API sends it both ways: some
		// products quote it ("50") and others do not (50), within the
		// same response. Declaring it as a string fails to decode the
		// whole product list the moment one unquoted value appears —
		// verified live, August 2026. Use Size.String() or Size.Int().
		Size shtypes.MaybeString `json:"size"`
	}

	// ListProductsResponse is the response for server/products.json.
	ListProductsResponse struct {
		Return []Product `json:"return"`
		models.APIResponse
	}
)

// UnmarshalJSON tolerates the empty-array form the API sends in place of
// an empty attributes object, and keeps unrecognised attributes in
// Extra.
//
// Some products come back with `"attributes": []` rather than `{}` —
// PHP's empty array serialising as a list — which would otherwise fail
// to decode into a struct.
func (a *ProductAttributes) UnmarshalJSON(data []byte) error {
	if shtypes.IsEmptyMapShape(data) {
		*a = ProductAttributes{}
		return nil
	}

	// Decode the known fields via an alias to avoid recursing.
	type plain ProductAttributes
	var known plain
	if err := json.Unmarshal(data, &known); err != nil {
		return fmt.Errorf("server: decoding product attributes: %w", err)
	}
	*a = ProductAttributes(known)

	// Then keep whatever is left.
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return fmt.Errorf("server: decoding product attributes: %w", err)
	}
	for _, k := range []string{"cores", "ram", "disk", "bandwidth", "partitions"} {
		delete(all, k)
	}
	if len(all) > 0 {
		a.Extra = all
	}
	return nil
}
