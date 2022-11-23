package models

import (
	"fmt"
	"net/http"
)

// APIResponse represents mutual API response values.
type APIResponse struct {
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// ErrorResponse reports the error caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Message  string         `json:"msg"`
	Status   bool           `json:"status"`
}

// Error returns a ErrorResponse message.
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}
