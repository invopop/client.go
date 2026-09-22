package invopop

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// responseError performs a request against a client that replies with the
// given status and body, returning the resulting error.
func responseError(t *testing.T, status int, body string) error {
	t.Helper()
	c := testClient(t, newRecorder(status, body))
	return c.get(context.Background(), "/test", new(Ping))
}

func TestResponseErrorHandle(t *testing.T) {
	t.Run("with a successful response", func(t *testing.T) {
		assert.NoError(t, responseError(t, http.StatusOK, `{}`))
	})

	t.Run("with a 201 response", func(t *testing.T) {
		assert.NoError(t, responseError(t, http.StatusCreated, `{}`))
	})

	t.Run("with an error response", func(t *testing.T) {
		err := responseError(t, http.StatusBadRequest, `{"message":"invalid request"}`)
		require.Error(t, err)
		assert.EqualError(t, err, "400: invalid request")
	})
}

func TestResponseErrorError(t *testing.T) {
	t.Run("with a code", func(t *testing.T) {
		err := responseError(t, http.StatusConflict, `{"code":"duplicate","message":"already exists"}`)
		require.Error(t, err)
		assert.EqualError(t, err, "409: (duplicate) already exists")
	})

	t.Run("without a code", func(t *testing.T) {
		err := responseError(t, http.StatusInternalServerError, `{"message":"boom"}`)
		require.Error(t, err)
		assert.EqualError(t, err, "500: boom")
	})

	t.Run("without a message", func(t *testing.T) {
		err := responseError(t, http.StatusBadGateway, `{}`)
		require.Error(t, err)
		assert.EqualError(t, err, "502: ")
	})
}

func TestResponseErrorAccessors(t *testing.T) {
	err := responseError(t, http.StatusNotFound, `{"code":"missing","message":"not here"}`)
	re := AsResponseError(err)
	require.NotNil(t, re)

	assert.Equal(t, http.StatusNotFound, re.StatusCode())
	assert.Equal(t, "missing", re.Code)
	assert.Equal(t, "not here", re.Message)
	require.NotNil(t, re.Response())
	assert.Equal(t, http.StatusNotFound, re.Response().StatusCode())
}

func TestResponseErrorFields(t *testing.T) {
	err := responseError(t, http.StatusUnprocessableEntity, `{
		"message": "validation failed",
		"fields": {
			"supplier": {"tax_id": "cannot be blank"}
		}
	}`)
	re := AsResponseError(err)
	require.NotNil(t, re)
	require.NotNil(t, re.Fields)
	assert.Equal(t, "cannot be blank", re.Fields.Get("supplier.tax_id").Message())
}

func TestAsResponseError(t *testing.T) {
	t.Run("with a response error", func(t *testing.T) {
		err := responseError(t, http.StatusNotFound, `{}`)
		assert.NotNil(t, AsResponseError(err))
	})

	t.Run("with a wrapped response error", func(t *testing.T) {
		err := fmt.Errorf("fetching entry: %w", responseError(t, http.StatusNotFound, `{}`))
		re := AsResponseError(err)
		require.NotNil(t, re)
		assert.Equal(t, http.StatusNotFound, re.StatusCode())
	})

	t.Run("with a regular error", func(t *testing.T) {
		assert.Nil(t, AsResponseError(errors.New("regular")))
	})

	t.Run("with a nil error", func(t *testing.T) {
		assert.Nil(t, AsResponseError(nil))
	})
}

func TestStatusHelpers(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		conflict  bool
		notFound  bool
		forbidden bool
	}{
		{name: "conflict", status: http.StatusConflict, conflict: true},
		{name: "not found", status: http.StatusNotFound, notFound: true},
		{name: "forbidden", status: http.StatusForbidden, forbidden: true},
		{name: "unauthorized matches none", status: http.StatusUnauthorized},
		{name: "server error matches none", status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := responseError(t, tt.status, `{}`)
			require.Error(t, err)
			assert.Equal(t, tt.conflict, IsConflict(err))
			assert.Equal(t, tt.notFound, IsNotFound(err))
			assert.Equal(t, tt.forbidden, IsForbidden(err))
		})
	}

	t.Run("with a regular error", func(t *testing.T) {
		err := errors.New("regular")
		assert.False(t, IsConflict(err))
		assert.False(t, IsNotFound(err))
		assert.False(t, IsForbidden(err))
	})

	t.Run("with a nil error", func(t *testing.T) {
		assert.False(t, IsConflict(nil))
		assert.False(t, IsNotFound(nil))
		assert.False(t, IsForbidden(nil))
	})
}

func TestErrAccessDenied(t *testing.T) {
	err := fmt.Errorf("%w: no token", ErrAccessDenied)
	assert.True(t, errors.Is(err, ErrAccessDenied))
}
