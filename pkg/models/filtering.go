package models

type (
	Filtering struct {
		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}
)
