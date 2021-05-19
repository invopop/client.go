package api

import (
	"net/http"
)

type APIError struct {
	Message string `json:"message"`
}

func (err *APIError) IsNil() bool {
	return err.Message == ""
}

type ClientError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func NewError(status int, msg string) *ClientError {
	return &ClientError{Status: status, Message: msg}
}

func NewInternalError(msg string) *ClientError {
	return &ClientError{Status: http.StatusInternalServerError, Message: msg}
}

func (e *ClientError) Error() string {
	return e.Message
}

func (e *ClientError) GetStatus() int {
	return e.Status
}

func FromError(err error) (s *ClientError, ok bool) {
	if err == nil {
		return nil, true
	}

	if se, ok := err.(interface {
		Error() string
		GetStatus() int
	}); ok {
		return &ClientError{Status: se.GetStatus(), Message: se.Error()}, true
	}

	return &ClientError{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	}, false
}
