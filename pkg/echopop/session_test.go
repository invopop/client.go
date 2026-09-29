package echopop_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/invopop/client.go/invopop"
	"github.com/invopop/client.go/pkg/echopop"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionEcho builds an echo instance with the cookie session middleware
// installed, matching what NewService does when given a session key.
func sessionEcho() *echo.Echo {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(echopop.GenerateCookieSecret()))))
	return e
}

func TestGetSessionWithValue(t *testing.T) {
	c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, echopop.GetSession(c), "no session in the context")

	c.Set("session", "not a session")
	assert.Nil(t, echopop.GetSession(c), "wrong type should not panic")
}

func TestLoadSession(t *testing.T) {
	ic := invopop.New()

	t.Run("with an authorization header", func(t *testing.T) {
		e := sessionEcho()
		var sess *invopop.Session
		e.GET("/", func(c echo.Context) error {
			sess = echopop.GetSession(c)
			return c.NoContent(http.StatusOK)
		}, echopop.LoadSession(ic))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, sess)
		assert.Equal(t, testToken, sess.Token)
		assert.NotNil(t, sess.Client())
	})

	t.Run("with a state query parameter", func(t *testing.T) {
		e := sessionEcho()
		var sess *invopop.Session
		e.GET("/", func(c echo.Context) error {
			sess = echopop.GetSession(c)
			return c.NoContent(http.StatusOK)
		}, echopop.LoadSession(ic))

		req := httptest.NewRequest(http.MethodGet, "/?state=state-token", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.NotNil(t, sess)
		assert.Equal(t, testStateToken, sess.Token)
	})

	t.Run("with no credentials provides an empty session", func(t *testing.T) {
		e := sessionEcho()
		var sess *invopop.Session
		e.GET("/", func(c echo.Context) error {
			sess = echopop.GetSession(c)
			return c.NoContent(http.StatusOK)
		}, echopop.LoadSession(ic))

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, sess, "a session should still be prepared")
		assert.Empty(t, sess.Token)
		assert.False(t, sess.Authorized())
	})
}

func TestStoreSessionCookie(t *testing.T) {
	ic := invopop.New()

	t.Run("stores and reloads a session", func(t *testing.T) {
		e := sessionEcho()

		// First request stores the session in a cookie.
		e.GET("/store", func(c echo.Context) error {
			sess := ic.Access().NewSession()
			sess.EnrollmentID = "enrollment-1"
			sess.OwnerID = "owner-1"
			sess.Token = testToken
			sess.TokenExpires = time.Now().Add(time.Hour).Unix()
			if err := echopop.StoreSessionCookie(c, sess); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})

		// Second request loads it back through the middleware.
		var loaded *invopop.Session
		e.GET("/load", func(c echo.Context) error {
			loaded = echopop.GetSession(c)
			return c.NoContent(http.StatusOK)
		}, echopop.LoadSession(ic))

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/store", nil))
		require.Equal(t, http.StatusOK, rec.Code)

		cookies := rec.Result().Cookies() //nolint:bodyclose
		require.NotEmpty(t, cookies, "a session cookie should be set")

		ck := cookies[0]
		assert.True(t, ck.HttpOnly, "should not be readable from AJAX")
		assert.True(t, ck.Secure)
		assert.Equal(t, http.SameSiteStrictMode, ck.SameSite, "should prevent CSRF")
		assert.Positive(t, ck.MaxAge)

		req := httptest.NewRequest(http.MethodGet, "/load", nil)
		req.AddCookie(ck)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req)

		require.Equal(t, http.StatusOK, rec2.Code)
		require.NotNil(t, loaded)
		assert.Equal(t, "enrollment-1", loaded.EnrollmentID)
		assert.Equal(t, "owner-1", loaded.OwnerID)
		assert.Equal(t, testToken, loaded.Token)
		assert.NotNil(t, loaded.Client(), "the client should be reattached")
	})

	t.Run("clears the cookie for an expired session", func(t *testing.T) {
		e := sessionEcho()
		e.GET("/", func(c echo.Context) error {
			sess := ic.Access().NewSession()
			sess.Token = testToken
			sess.TokenExpires = time.Now().Add(-time.Hour).Unix()
			if err := echopop.StoreSessionCookie(c, sess); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		cookies := rec.Result().Cookies() //nolint:bodyclose
		require.NotEmpty(t, cookies)
		assert.Negative(t, cookies[0].MaxAge, "an expired session should delete the cookie")
	})

	t.Run("never sets a negative max age", func(t *testing.T) {
		e := sessionEcho()
		e.GET("/", func(c echo.Context) error {
			// Expiry in the future by less than a second: the computed
			// MaxAge can round to zero but must not go negative.
			sess := ic.Access().NewSession()
			sess.Token = testToken
			sess.TokenExpires = time.Now().Unix() + 1
			if err := echopop.StoreSessionCookie(c, sess); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		cookies := rec.Result().Cookies() //nolint:bodyclose
		require.NotEmpty(t, cookies)
		assert.GreaterOrEqual(t, cookies[0].MaxAge, 0)
	})

	t.Run("with a corrupt session cookie", func(t *testing.T) {
		e := sessionEcho()
		e.GET("/store", func(c echo.Context) error {
			cs, err := session.Get("_echopop", c)
			require.NoError(t, err)
			// The middleware expects a JSON string under this key.
			cs.Values["session"] = "{not json"
			require.NoError(t, cs.Save(c.Request(), c.Response()))
			return c.NoContent(http.StatusOK)
		})
		e.GET("/load", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		}, echopop.LoadSession(ic))

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/store", nil))
		cookies := rec.Result().Cookies() //nolint:bodyclose
		require.NotEmpty(t, cookies)

		req := httptest.NewRequest(http.MethodGet, "/load", nil)
		req.AddCookie(cookies[0])
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req)

		assert.Equal(t, http.StatusInternalServerError, rec2.Code,
			"an unparseable session should surface as an error")
	})
}

func TestSessionCookieRoundTripsMeta(t *testing.T) {
	ic := invopop.New()
	e := sessionEcho()

	e.GET("/store", func(c echo.Context) error {
		sess := ic.Access().NewSession()
		sess.EnrollmentID = "enrollment-1"
		sess.Token = testToken
		sess.TokenExpires = time.Now().Add(time.Hour).Unix()
		sess.Meta = map[string]string{"workspace": "acme"}
		sess.Data = json.RawMessage(`{"setting":true}`)
		sess.RedirectURI = "https://example.com/back"
		return echopop.StoreSessionCookie(c, sess)
	})

	var loaded *invopop.Session
	e.GET("/load", func(c echo.Context) error {
		loaded = echopop.GetSession(c)
		return c.NoContent(http.StatusOK)
	}, echopop.LoadSession(ic))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/store", nil))
	cookies := rec.Result().Cookies() //nolint:bodyclose
	require.NotEmpty(t, cookies)

	req := httptest.NewRequest(http.MethodGet, "/load", nil)
	req.AddCookie(cookies[0])
	e.ServeHTTP(httptest.NewRecorder(), req)

	require.NotNil(t, loaded)
	assert.Equal(t, map[string]string{"workspace": "acme"}, loaded.Meta)
	assert.JSONEq(t, `{"setting":true}`, string(loaded.Data))
	assert.Equal(t, "https://example.com/back", loaded.RedirectURI)
}
