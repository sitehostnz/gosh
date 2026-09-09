package integrated

type (
	// ListRequest lists the integrated stacks on a server.
	ListRequest struct {
		ServerName string `url:"server_name"`
	}

	// AddRequest creates an integrated stack on a server.
	//
	// Name is the stack name, which is also its image code — see the
	// Stack* constants. Nothing else is configurable: the platform
	// generates the compose file, so there is no label, no SSL flag and
	// no compose to pass, unlike cloud/stack/add.json.
	AddRequest struct {
		ServerName string `url:"server"`
		Name       string `url:"name"`
	}
)
