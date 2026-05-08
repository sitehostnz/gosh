package volume

type (
	// ListOptions represents optional filters for the list_all
	// call. All fields are optional.
	ListOptions struct {
		Name       string `url:"filters[name],omitempty"`
		ServerName string `url:"filters[server_name],omitempty"`
		Container  string `url:"filters[container],omitempty"`
		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}

	// GetOptions identifies the volume to fetch.
	//
	// Note: get and delete use the unprefixed parameter names
	// (server / volume) — distinct from add / mount /
	// update_mounts which use server_name / volume_name. This
	// reflects the API's actual surface; tags are explicit.
	GetOptions struct {
		Server string `url:"server"`
		Volume string `url:"volume"`
	}

	// DeleteOptions identifies the volume to delete. Same
	// parameter naming as GetOptions.
	DeleteOptions struct {
		Server string `url:"server"`
		Volume string `url:"volume"`
	}

	// AddOptions describes a new volume to create. ContainerNames
	// is an optional list of container identifiers to attach the
	// volume to at creation time.
	AddOptions struct {
		ServerName     string
		VolumeName     string
		ContainerNames []string
	}

	// ContainerMount describes a single container target for
	// volume mount / unmount operations: the stack the container
	// belongs to and the container's name.
	ContainerMount struct {
		StackName     string
		ContainerName string
	}

	// MountOptions describes a volume to mount and the containers
	// to mount it to.
	MountOptions struct {
		ServerName string
		VolumeName string
		Containers []ContainerMount
	}

	// UpdateMountsOptions describes incremental mount changes —
	// containers to attach (Add) and detach (Remove). Either or
	// both may be set.
	UpdateMountsOptions struct {
		ServerName string
		VolumeName string
		Add        []ContainerMount
		Remove     []ContainerMount
	}
)
