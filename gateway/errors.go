package gateway

import (
	"errors"
	"fmt"
)

// Error adds to the standard protocol Error type a standard response
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code.String(), e.Message)
}

// AsError returns an Error type from an error, if it is one.
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// IsInternalError returns true if the error is internal
func IsInternalError(err error) bool {
	if e := AsError(err); e != nil {
		return e.Code == ErrorCode_INTERNAL
	}
	return false
}

// IsValidationError returns true if the error is related to data validation.
func IsValidationError(err error) bool {
	if e := AsError(err); e != nil {
		return e.Code == ErrorCode_INVALID
	}
	return false
}

// IsNotFoundError returns true if something was not found.
func IsNotFoundError(err error) bool {
	if e := AsError(err); e != nil {
		return e.Code == ErrorCode_NOT_FOUND
	}
	return false
}
