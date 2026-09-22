package echopop_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/invopop/client.go/pkg/echopop"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Shared fixture values, kept as constants so that the repeated literals
// stay consistent across the package's tests.
const (
	testToken      = "my-token"
	testStateToken = "state-token"
)

// newContext builds an echo context around a request and recorder.
func newContext(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec), rec
}

func TestAuthToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		query  string
		want   string
	}{
		{
			name:   "with a bearer token",
			header: "Bearer my-token",
			want:   testToken,
		},
		{
			name:   "with a lowercase bearer prefix",
			header: "bearer my-token",
			want:   testToken,
		},
		{
			name:   "with a mixed case bearer prefix",
			header: "BeArEr my-token",
			want:   testToken,
		},
		{
			name:   "with a non-bearer scheme",
			header: "Basic dXNlcjpwYXNz",
			want:   "",
		},
		{
			name:   "with a short header",
			header: "Bearer",
			want:   "",
		},
		{
			name:  "with a state query parameter",
			query: testStateToken,
			want:  testStateToken,
		},
		{
			name:   "prefers the header over the query parameter",
			header: "Bearer header-token",
			query:  testStateToken,
			want:   "header-token",
		},
		{
			name:   "with an empty bearer token falls back to the query",
			header: "Bearer ",
			query:  testStateToken,
			want:   testStateToken,
		},
		{
			name: "with nothing",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/"
			if tt.query != "" {
				target += "?state=" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			c, _ := newContext(req)
			assert.Equal(t, tt.want, echopop.AuthToken(c))
		})
	}
}

func TestGenerateCookieSecret(t *testing.T) {
	s1 := echopop.GenerateCookieSecret()
	s2 := echopop.GenerateCookieSecret()

	assert.Len(t, s1, 64, "32 random bytes as hex")
	assert.Regexp(t, "^[0-9a-f]{64}$", s1)
	assert.NotEqual(t, s1, s2, "should be random")
}

func TestRender(t *testing.T) {
	t.Run("renders the component", func(t *testing.T) {
		c, rec := newContext(httptest.NewRequest(http.MethodGet, "/", nil))

		comp := templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
			_, err := w.Write([]byte("<h1>Hello</h1>"))
			return err
		})

		require.NoError(t, echopop.Render(c, http.StatusOK, comp))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "<h1>Hello</h1>", rec.Body.String())
	})

	t.Run("honours the status code", func(t *testing.T) {
		c, rec := newContext(httptest.NewRequest(http.MethodGet, "/", nil))

		comp := templ.ComponentFunc(func(_ context.Context, _ io.Writer) error { return nil })
		require.NoError(t, echopop.Render(c, http.StatusCreated, comp))
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("with a failing component", func(t *testing.T) {
		c, _ := newContext(httptest.NewRequest(http.MethodGet, "/", nil))

		comp := templ.ComponentFunc(func(_ context.Context, _ io.Writer) error {
			return errors.New("render failed")
		})

		err := echopop.Render(c, http.StatusOK, comp)
		require.Error(t, err)

		var he *echo.HTTPError
		require.True(t, errors.As(err, &he))
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Equal(t, "render failed", he.Message)
	})
}

func TestNewService(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		svc := echopop.NewService()
		require.NotNil(t, svc)

		var e *echo.Echo
		svc.Serve(func(inner *echo.Echo) { e = inner })
		assert.NotNil(t, e)
	})

	t.Run("with a cookie session key", func(t *testing.T) {
		svc := echopop.NewService(
			echopop.WithCookieSessionKey(echopop.GenerateCookieSecret()),
		)
		require.NotNil(t, svc)
	})

	t.Run("root provides a group", func(t *testing.T) {
		svc := echopop.NewService()
		called := false
		svc.Root(func(g *echo.Group) {
			called = true
			require.NotNil(t, g)
			g.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})
		})
		assert.True(t, called)
	})

	t.Run("serves registered routes", func(t *testing.T) {
		svc := echopop.NewService()
		var e *echo.Echo
		svc.Serve(func(inner *echo.Echo) {
			e = inner
			inner.GET("/ping", func(c echo.Context) error {
				return c.String(http.StatusOK, "pong")
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "pong", rec.Body.String())
	})

	t.Run("recovers from a panicking handler", func(t *testing.T) {
		svc := echopop.NewService()
		var e *echo.Echo
		svc.Serve(func(inner *echo.Echo) {
			e = inner
			inner.GET("/panic", func(_ echo.Context) error {
				panic("boom")
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() { e.ServeHTTP(rec, req) })
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestServiceStartStop(t *testing.T) {
	svc := echopop.NewService()

	errs := make(chan error, 1)
	go func() { errs <- svc.Start("0") }() // port 0 picks a free port

	// Shutdown should be clean, and Start should not report the resulting
	// ErrServerClosed as a failure.
	require.Eventually(t, func() bool {
		return svc.Stop(context.Background()) == nil
	}, 5*time.Second, 10*time.Millisecond)

	select {
	case err := <-errs:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the server to stop")
	}
}

func TestServiceStaticRootFS(t *testing.T) {
	svc := echopop.NewService()
	svc.StaticRootFS(testAssets, "testdata/assets")

	var e *echo.Echo
	svc.Serve(func(inner *echo.Echo) { e = inner })

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "console.log")
}
