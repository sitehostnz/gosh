package version

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListAllResponse is the return from
	// /cloud/image/version/list_all.json.
	ListAllResponse struct {
		Return struct {
			models.Pagination
			Versions []Version `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// Build is the body of GetBuildResponse — status plus the raw
	// CI build trace.
	Build struct {
		BuildStatus string `json:"build_status"`
		BuildTrace  string `json:"build_trace"`
	}

	// GetBuildResponse is the return from
	// /cloud/image/version/get_build.json.
	GetBuildResponse struct {
		Return Build `json:"return"`
		models.APIResponse
	}

	// JobResponse wraps the scheduler-job return for write
	// operations on this namespace (currently only Delete).
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
