package invopop

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

	// Key is the short error key provided by the server (e.g. "validation").
	Key string `json:"key,omitempty"`

	// Message contains a human readable response message from the API in the case
	// of an error.
	Message string `json:"message,omitempty"`

	// Faults provides the list of individual faults detected by the server,
	// typically for validation errors.
	Faults []*Fault `json:"faults,omitempty"`

	// Fields provides a nested map of field-level error messages.
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
	parts := []string{fmt.Sprintf("%d", r.response.StatusCode())}
	if r.Key != "" {
		parts = append(parts, r.Key)
	}
	if r.Message != "" {
		parts = append(parts, r.Message)
	}
	if len(r.Faults) > 0 {
		fs := make([]string, 0, len(r.Faults))
		for _, f := range r.Faults {
			fs = append(fs, f.describe())
		}
		parts = append(parts, strings.Join(fs, "; "))
	}
	if len(parts) == 1 {
		if body := strings.TrimSpace(r.response.String()); body != "" {
			parts = append(parts, body)
		}
	}
	return strings.Join(parts, ": ")
}

// describe returns a compact human-readable representation of the fault, used
// by ResponseError.Error to surface fault details in error strings.
func (f *Fault) describe() string {
	switch {
	case f.Code != "" && f.Message != "":
		return fmt.Sprintf("%s: %s", f.Code, f.Message)
	case f.Message != "":
		return f.Message
	case f.Code != "":
		return f.Code
	default:
		return ""
	}
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
