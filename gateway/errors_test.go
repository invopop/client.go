package gateway

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "internal",
			err:  &Error{Code: ErrorCode_INTERNAL, Message: "boom"},
			want: "INTERNAL: boom",
		},
		{
			name: "invalid",
			err:  &Error{Code: ErrorCode_INVALID, Message: "bad data"},
			want: "INVALID: bad data",
		},
		{
			name: "not found",
			err:  &Error{Code: ErrorCode_NOT_FOUND, Message: "missing"},
			want: "NOT_FOUND: missing",
		},
		{
			name: "empty message",
			err:  &Error{Code: ErrorCode_INVALID},
			want: "INVALID: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestAsError(t *testing.T) {
	t.Run("with a gateway error", func(t *testing.T) {
		src := &Error{Code: ErrorCode_INVALID, Message: "bad"}
		out := AsError(error(src))
		require.NotNil(t, out)
		assert.Equal(t, src, out)
	})

	t.Run("with a wrapped gateway error", func(t *testing.T) {
		src := &Error{Code: ErrorCode_NOT_FOUND, Message: "missing"}
		out := AsError(fmt.Errorf("fetching: %w", error(src)))
		require.NotNil(t, out)
		assert.Equal(t, src, out)
	})

	t.Run("with a regular error", func(t *testing.T) {
		assert.Nil(t, AsError(errors.New("regular")))
	})

	t.Run("with a nil error", func(t *testing.T) {
		assert.Nil(t, AsError(nil))
	})
}

func TestErrorCodeHelpers(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		internal  bool
		validaton bool
		notFound  bool
	}{
		{
			name:     "internal error",
			err:      error(&Error{Code: ErrorCode_INTERNAL}),
			internal: true,
		},
		{
			name:      "validation error",
			err:       error(&Error{Code: ErrorCode_INVALID}),
			validaton: true,
		},
		{
			name:     "not found error",
			err:      error(&Error{Code: ErrorCode_NOT_FOUND}),
			notFound: true,
		},
		{
			name:     "wrapped not found error",
			err:      fmt.Errorf("wrapped: %w", error(&Error{Code: ErrorCode_NOT_FOUND})),
			notFound: true,
		},
		{
			name: "regular error",
			err:  errors.New("regular"),
		},
		{
			name: "nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.internal, IsInternalError(tt.err))
			assert.Equal(t, tt.validaton, IsValidationError(tt.err))
			assert.Equal(t, tt.notFound, IsNotFoundError(tt.err))
		})
	}
}
