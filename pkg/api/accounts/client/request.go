package client

// ListSubAccountsRequest selects + paginates the sub-accounts list.
// All fields are optional; zero values mean "no filter / API
// default".
type ListSubAccountsRequest struct {
	// Name partial-matches against the sub-account display name.
	Name string
	// IncludeClosed includes sub-accounts whose Closed date is set.
	IncludeClosed bool
	// SortBy is the field name to sort by ("name", "joined", etc.).
	SortBy string
	// SortDir is "ASC" or "DESC".
	SortDir string
	// PageSize and PageNumber control pagination.
	PageSize   int
	PageNumber int
}
