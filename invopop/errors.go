package invopop

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
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

// IsConflict is a helper that will provide the response error object
// if the error is a conflict.
func IsConflict(err error) *ResponseError {
	return isError(err, http.StatusConflict)
}

// IsNotFound returns the error response if the status is not found.
func IsNotFound(err error) *ResponseError {
	return isError(err, http.StatusNotFound)
}

// IsForbidden returns the error response if the status is forbidden.
func IsForbidden(err error) *ResponseError {
	return isError(err, http.StatusForbidden)
}

func isError(err error, status int) *ResponseError {
	var re *ResponseError
	if errors.As(err, &re) {
		if re.StatusCode() == status {
			return re
		}
	}
	return nil
}
