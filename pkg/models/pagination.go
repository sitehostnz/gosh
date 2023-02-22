package models

type (
	// Pagination represents the pagination information return by the API.
	Pagination struct {
		TotalItems   int `json:"total_items"`
		CurrentItems int `json:"current_items"`
		CurrentPage  int `json:"current_page"`
		TotalPages   int `json:"total_pages"`
	}
)
