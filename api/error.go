package api

// APIError defines the common API response for errors.
type APIError struct {
	Message string `json:"message"`
}

// Error defines a general error structure.
type Error struct {
	Message string `json:"message"`
}

// NewError create a new instance of Error from a given `code` and `msg`.
func NewError(msg string) *Error {
	return &Error{Message: msg}
}

// Error returns a string message of the error.
func (e *Error) Error() string {
	return e.Message
}
