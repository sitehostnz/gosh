package image

type (
	// GetRequest represents a request to get a specific image from
	// the /cloud/image/get.json endpoint.
	GetRequest struct {
		Code string `json:"code"`
	}

	// CreateRequest creates a new custom image via
	// /cloud/image/create.json. Label is required. Code is optional —
	// the API generates one from the label if omitted. ForkID is the
	// id of a public SiteHost image to fork from (optional; omit to
	// build from scratch). SSHKeys is a list of customer-level SSH
	// key IDs to grant access to the backing GitLab repository.
	CreateRequest struct {
		Label   string
		Code    string
		ForkID  int
		SSHKeys []int
	}

	// DeleteRequest deletes a custom image via
	// /cloud/image/delete.json.
	DeleteRequest struct {
		Code string
	}

	// GetChangelogRequest fetches the change log for a public
	// SiteHost image via /cloud/image/get_changelog.json. Code is the
	// public image's code (e.g. "sitehost-php55").
	GetChangelogRequest struct {
		Code string
	}
)
