package token

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// Summary is a token as list_all reports it.
	Summary struct {
		KeyID       string `json:"key_id"`
		Label       string `json:"label"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// Token is a token as get reports it: the summary plus its
	// configuration. The secret is never included; see AddResponse.
	Token struct {
		Summary
		Config Config `json:"config"`
	}

	// Config is a token's access configuration.
	//
	// RequestsPerMinute and the CostLimit* fields are sent by the API
	// but cannot be set through it: add and update accept neither. They
	// have only been observed as 0, which is taken to mean "no limit".
	// The cost limits are declared float64 because they are NZD amounts;
	// a whole-number 0 decodes into float64 without complaint.
	Config struct {
		AllowedIPs        []string `json:"allowed_ips"`
		Capabilities      []string `json:"capabilities"`
		RequestsPerMinute int      `json:"requests_per_minute"`
		CostLimitHourly   float64  `json:"cost_limit_hourly"`
		CostLimitDaily    float64  `json:"cost_limit_daily"`
		CostLimitMonthly  float64  `json:"cost_limit_monthly"`
	}

	// AddResponse represents the result of issuing a token.
	AddResponse struct {
		Return struct {
			KeyID string `json:"key_id"`
			Label string `json:"label"`

			// Token is the secret, used as a Bearer token against
			// ai.sitehost.nz. This is the only time the API returns
			// it.
			Token string `json:"token"`
		} `json:"return"`
		models.APIResponse
	}

	// GetResponse represents a single token's details.
	GetResponse struct {
		Return Token `json:"return"`
		models.APIResponse
	}

	// ListResponse represents the listing of the account's tokens.
	//
	// Revoked tokens are not listed, though the platform keeps their
	// records for billing.
	ListResponse struct {
		Return struct {
			models.Pagination
			Tokens []Summary `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateResponse represents the result of changing a token.
	//
	// Updated is true whenever the request was accepted, including when
	// it changed nothing; it is not evidence that a value was applied.
	UpdateResponse struct {
		Return struct {
			Updated bool `json:"updated"`
		} `json:"return"`
		models.APIResponse
	}

	// DeleteResponse represents the result of revoking a token.
	DeleteResponse struct {
		Return struct {
			Deleted bool `json:"deleted"`
		} `json:"return"`
		models.APIResponse
	}
)
