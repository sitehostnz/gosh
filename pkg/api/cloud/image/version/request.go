package version

type (
	// ListAllRequest lists versions of a custom image. ImageID is the
	// numeric id of the image (from cloud/image/get's "id" field, or
	// list_all). Filters are optional pagination/ordering knobs.
	ListAllRequest struct {
		ImageID    int
		SortBy     string
		SortDir    string
		PageSize   int
		PageNumber int
	}

	// GetBuildRequest fetches the build log for a specific build.
	// BuildID is the numeric build ID (Version.BuildID from
	// ListAllResponse).
	GetBuildRequest struct {
		Code    string
		BuildID string
	}

	// DeleteRequest removes a specific version of a custom image.
	// Version is the version string (e.g. "1.1-1076"), available
	// from ListAllResponse.
	DeleteRequest struct {
		Code    string
		Version string
	}
)
