package invopop

import (
	"errors"
	"fmt"
	"net/http"

	"resty.dev/v3"
)

// ResponseError is a wrapper around error responses from the server that will handle
// error messages.
type ResponseError struct {
	response *resty.Response

	// Code is the error code which may have been provided by the server.
	Code string `json:"code"`

	// Message contains a human readable response message from the API in the case
	// of an error.
	Message string `json:"message"`

	// Fields provides a nested map of
	Fields *Dict `json:"fields,omitempty"`
}

// handle will wrap the resty response to provide our own Response object that
// wraps around any errors that might have happened with the connection or response.
func (r *ResponseError) handle(res *resty.Response) error {
	if res.IsSuccess() {
		return nil
	}
	r.response = res
	return r
}

// StatusCode provides the response status code, or 0 if an error occurred.
func (r *ResponseError) StatusCode() int {
	return r.response.StatusCode()
}

// Error provides the response error string.
func (r *ResponseError) Error() string {
	if r.Code != "" {
		return fmt.Sprintf("%d: (%s) %s", r.response.StatusCode(), r.Code, r.Message)
	}
	return fmt.Sprintf("%d: %v", r.response.StatusCode(), r.Message)
}

// Response provides underlying response, in case it might be useful for
// debugging.
func (r *ResponseError) Response() *resty.Response {
	return r.response
}

// IsConflict is a helper that will return true if the response error is
// a conflict.
func IsConflict(err error) bool {
	return asError(err, http.StatusConflict) != nil
}

// IsNotFound is a helper that will return true if the response error is
// a not found.
func IsNotFound(err error) bool {
	return asError(err, http.StatusNotFound) != nil
}

// IsForbidden is a helper that will return true if the response error is
// a forbidden.
func IsForbidden(err error) bool {
	return asError(err, http.StatusForbidden) != nil
}

// AsResponseError will extract the ResponseError from the provided error
// or return nil if no match found.
func AsResponseError(err error) *ResponseError {
	var re *ResponseError
	if errors.As(err, &re) {
		return re
	}
	return nil
}

func asError(err error, status int) *ResponseError {
	re := AsResponseError(err)
	if re != nil && re.StatusCode() == status {
		return re
	}
	return nil
}
