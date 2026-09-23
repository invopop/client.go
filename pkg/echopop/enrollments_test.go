package echopop_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/invopop/client.go/invopop"
	"github.com/invopop/client.go/pkg/echopop"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apiServer starts a stub Invopop API replying with the given status and body,
// and returns a client pointed at it with OAuth credentials configured.
func apiServer(t *testing.T, status int, body string) *invopop.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	ic := invopop.New(invopop.WithConfig(&invopop.Config{
		BaseURL:      srv.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}))
	return ic
}

func TestLoadEnrollment(t *testing.T) {
	t.Run("without an auth token", func(t *testing.T) {
		ic := apiServer(t, http.StatusOK, `{}`)
		c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))

		err := echopop.LoadEnrollment(ic, c)
		require.Error(t, err)
		assert.ErrorIs(t, err, invopop.ErrAccessDenied)
		assert.Contains(t, err.Error(), "missing auth token")
	})

	t.Run("with a valid token", func(t *testing.T) {
		ic := apiServer(t, http.StatusOK, `{
			"id": "enrollment-1",
			"owner_id": "owner-1",
			"token": "enrollment-token",
			"token_expires": 9999999999
		}`)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		c, _ := newContext(req)

		require.NoError(t, echopop.LoadEnrollment(ic, c))

		en := echopop.GetEnrollment(c)
		require.NotNil(t, en)
		assert.Equal(t, "enrollment-1", en.ID)
		assert.Equal(t, "owner-1", en.OwnerID)

		assert.NotNil(t, echopop.GetClient(c), "a prepared client should be in the context")
	})

	t.Run("with a state query parameter", func(t *testing.T) {
		ic := apiServer(t, http.StatusOK, `{"id":"enrollment-1","token":"tok"}`)
		c, _ := newContext(httptest.NewRequest(http.MethodGet, "/?state=my-token", nil))

		require.NoError(t, echopop.LoadEnrollment(ic, c))
		require.NotNil(t, echopop.GetEnrollment(c))
	})

	t.Run("with an unknown enrollment", func(t *testing.T) {
		ic := apiServer(t, http.StatusNotFound, `{"message":"not found"}`)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		c, _ := newContext(req)

		err := echopop.LoadEnrollment(ic, c)
		require.Error(t, err)
		assert.ErrorIs(t, err, invopop.ErrAccessDenied)
		assert.Contains(t, err.Error(), "enrollment not found")
	})

	t.Run("with a server error", func(t *testing.T) {
		ic := apiServer(t, http.StatusInternalServerError, `{"message":"boom"}`)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		c, _ := newContext(req)

		err := echopop.LoadEnrollment(ic, c)
		require.Error(t, err)
		assert.False(t, errors.Is(err, invopop.ErrAccessDenied))
	})
}

func TestAuthEnrollment(t *testing.T) {
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	t.Run("calls the next handler when authorized", func(t *testing.T) {
		ic := apiServer(t, http.StatusOK, `{"id":"enrollment-1","token":"tok"}`)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		c, rec := newContext(req)

		require.NoError(t, echopop.AuthEnrollment(ic)(handler)(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("returns unauthorized when access is denied", func(t *testing.T) {
		ic := apiServer(t, http.StatusOK, `{}`)
		c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))

		err := echopop.AuthEnrollment(ic)(handler)(c)
		require.Error(t, err)

		var he *echo.HTTPError
		require.True(t, errors.As(err, &he))
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("returns an internal error for other failures", func(t *testing.T) {
		ic := apiServer(t, http.StatusInternalServerError, `{"message":"boom"}`)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		c, _ := newContext(req)

		err := echopop.AuthEnrollment(ic)(handler)(c)
		require.Error(t, err)

		var he *echo.HTTPError
		require.True(t, errors.As(err, &he))
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Error(t, he.Internal)
	})
}

func TestGetEnrollmentWithValue(t *testing.T) {
	c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, echopop.GetEnrollment(c))

	// A value of the wrong type should not panic.
	c.Set("enrollment", "not an enrollment")
	assert.Nil(t, echopop.GetEnrollment(c))
}

func TestGetClientWithValue(t *testing.T) {
	c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, echopop.GetClient(c))

	c.Set("invopop-client", "not a client")
	assert.Nil(t, echopop.GetClient(c))
}
