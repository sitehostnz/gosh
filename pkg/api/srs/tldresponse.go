package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// TLD describes a supported TLD's properties.
	TLD struct {
		TLD         string `json:"tld"`
		Term        string `json:"term"`
		Type        string `json:"type"`
		CanTransfer string `json:"can_transfer"`
		CanProtect  string `json:"can_protect"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// ListValidTLDsResponse represents the response from list_valid_tlds.
	ListValidTLDsResponse struct {
		Return []TLD `json:"return"`
		models.APIResponse
	}

	// PricingTier describes a single pricing tier (count threshold,
	// price at that tier).
	PricingTier struct {
		Type     int    `json:"type"`
		TypeName string `json:"type_name"`
		Count    string `json:"count"`
		Price    string `json:"price"`
	}

	// GetPricingTiersResponse represents the response from get_pricing_tiers.
	GetPricingTiersResponse struct {
		Return []PricingTier `json:"return"`
		models.APIResponse
	}
)
