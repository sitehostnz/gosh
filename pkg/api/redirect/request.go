package redirect

// ListRedirectsRequest paginates / sorts the redirect listing. All
// fields are optional.
type ListRedirectsRequest struct {
	SortBy     string
	SortDir    string
	PageSize   int
	PageNumber int
}
