package server

type (
	// GetUpdateWindowRequest identifies the CCS whose update-window
	// configuration to read.
	GetUpdateWindowRequest struct {
		ServerName string `url:"server_name"`
	}

	// SetUpdateWindowRequest configures the maintenance / patching
	// window for a CCS. All fields are required by the API.
	//
	// DayOfWeek is 1–7 (1 = Monday); HourOfDay is 0–23;
	// MinuteOfHour is 0–59. Enabled controls whether the window is
	// active (false disables auto-maintenance).
	SetUpdateWindowRequest struct {
		ServerName   string `url:"server_name"`
		Enabled      int    `url:"enabled"` // 0 or 1
		DayOfWeek    int    `url:"day_of_week"`
		HourOfDay    int    `url:"hour_of_day"`
		MinuteOfHour int    `url:"minute_of_hour"`
	}

	// UpdateMinimumTLSVersionRequest sets the minimum TLS version
	// the CCS's nginx-proxy will negotiate. The API accepts the
	// "TLSv1.x" format (e.g. "TLSv1.1", "TLSv1.2", "TLSv1.3");
	// values like "1.2" or "TLS_1_2" are rejected.
	//
	// Note: there's no corresponding read endpoint. Consumers must
	// either track the value they set or observe via TLS handshake
	// (see examples/probe-tls-default).
	UpdateMinimumTLSVersionRequest struct {
		ServerName        string `url:"server_name"`
		MinimumTLSVersion string `url:"minimum_tls_version"`
	}
)
