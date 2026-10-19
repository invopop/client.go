package invopop

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"resty.dev/v3"
)

// Common errors directly exposed by the invopop package and not considered
// response errors.
var (
	ErrAccessDenied = errors.New("access denied")
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

// maxErrorBodySize bounds how much of an error response is read when the
// response body is otherwise streamed.
const maxErrorBodySize = 64 * 1024

// rawResponseError reads the error from a response whose body was left
// unparsed, as for a download, and closes it: resty only decodes errors when
// it parses the response. A body that is not an API error, such as another
// server's error page, leaves the status text as the message.
func rawResponseError(res *resty.Response) error {
	if res.IsSuccess() {
		return nil
	}
	defer res.Body.Close() //nolint:errcheck
	re := new(ResponseError)
	if data, err := io.ReadAll(io.LimitReader(res.Body, maxErrorBodySize)); err == nil {
		_ = json.Unmarshal(data, re)
	}
	if re.Message == "" {
		re.Message = http.StatusText(res.StatusCode())
	}
	return re.handle(res)
}
