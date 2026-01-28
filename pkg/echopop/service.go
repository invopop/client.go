package echopop

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/foolin/goview"
	echoview "github.com/foolin/goview/supports/echoview-v4"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

// Service provides a wrapper around Echo that makes it a bit easier
// to start up a new service that will provide an HTTP server.
type Service struct {
	echo       *echo.Echo
	sessionKey string
}

// Option defines a configuration option for the Service.
type Option func(s *Service)

// WithCookieSessionKey sets the session key to be used by the service. This must be
// a sufficiently long random string to ensure the security of the session cookies,
// with a minimum length of 32 bytes recommended. See the GenerateCookieSecret function
// for a way to generate a suitable random string.
func WithCookieSessionKey(key string) Option {
	return func(s *Service) {
		s.sessionKey = key
	}
}

// NewService instantiates a new echo service using some reasonable
// defaults. Typical usage example:
//
//	svc := echopop.NewService(
//	    echopop.WithCookieSessionKey("your-randomly-long-secret"),
//	)
//	svc.Serve(func(e *echo.Echo) {
//	  e.StaticFS("/", assets.Content)
//	  g := e.Group("/api", svc.LoadSession())
//	  g.GET("/test", testHandler)
//	})
func NewService(opts ...Option) *Service {
	s := &Service{
		echo: echo.New(),
	}
	for _, opt := range opts {
		opt(s)
	}

	s.echo.Use(logRequest())
	s.echo.Use(middleware.Recover())

	if s.sessionKey != "" {
		s.echo.Use(
			session.Middleware(sessions.NewCookieStore([]byte(s.sessionKey))),
		)
	}

	return s
}

// Serve provides the echo instance in a callback method which
// might be semantically useful, but doesn't do much.
func (s *Service) Serve(fn func(*echo.Echo)) {
	fn(s.echo)
}

// Root provides an Echo Group instance from which to serve HTTP
// requests.
func (s *Service) Root(fn func(*echo.Group)) {
	s.Serve(func(e *echo.Echo) {
		fn(e.Group(""))
	})
}

// StaticRootFS is used to serve static file assets from the web folder root.
// The root is the path to the folder in the filesystem, and the fs is the
// filesystem object to use.
func (s *Service) StaticRootFS(fs fs.FS, root string) {
	s.echo.StaticFS("/", echo.MustSubFS(fs, root))
}

// AuthToken is a convenience method to extract the authentication token from
// the request context. It will look for a Bearer token in the Authorization
// header, or a "state" query parameter which is often used in OAuth 2.0 flows.
// If no token is found, an empty string is returned.
func AuthToken(c echo.Context) string {
	tok := ""
	auth := c.Request().Header.Get("Authorization")
	if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
		tok = auth[7:]
	}
	if tok == "" {
		// try to use OAuth 2.0 state query param
		tok = c.QueryParam(enrollmentStateKey)
	}
	return tok
}

// Render will prepare the echo templating feature using "goview"
// and the recommended defaults for modules.
//
// If the source file path is not found, it'll attempt to use the
// provided filesystem object. Assets served from disk will have
// caching disabled to facilitate rapid reloading.
//
// Usage example:
//
//	m.Render("templates", "./assets", assets.Content)
//
// Where the name ("templates") defines the path inside the source assets
// folder ("./assets") to find the data, *or* if not available, use
// the `assets.Content` embedded filesystem.
//
// Deprecated: use the echopop.Render method that uses templ.
func (s *Service) Render(name, src string, fs embed.FS) {
	var ev *echoview.ViewEngine
	base := path.Join(src, name)
	_, err := os.Stat(base)
	if errors.Is(err, os.ErrNotExist) {
		// No files found in path, use the embedded FS
		ev = echoview.New(goview.Config{
			Root:      name,
			Extension: ".html",
			Master:    "layouts/main",
		})
		ev.SetFileHandler(prepareFileHandler(fs))
	} else {
		ev = echoview.New(goview.Config{
			Root:         base,
			Extension:    ".html",
			Master:       "layouts/main",
			DisableCache: true,
		})
	}
	s.echo.Renderer = ev
}

func prepareFileHandler(fs embed.FS) goview.FileHandler {
	return func(config goview.Config, tplFile string) (string, error) {
		p := path.Join(config.Root, tplFile+config.Extension)
		data, err := fs.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("ViewEngine render read name:%v, path:%v, error: %v", tplFile, p, err)
		}
		return string(data), nil
	}
}

// Start will start the service on the provided port in the foreground.
func (s *Service) Start(port string) error {
	if err := s.echo.Start(":" + port); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

// Stop will gracefully stop the service with the passed context that probably
// includes a timeout.
func (s *Service) Stop(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// GenerateCookieSecret will generate a sufficiently random secret
// suitable for use with sessions.
func GenerateCookieSecret() string {
	key := securecookie.GenerateRandomKey(32)
	return fmt.Sprintf("%x", key)
}

func logRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tn := time.Now()
			req := c.Request()
			res := c.Response()
			err := next(c)
			if err != nil {
				// Required to log the actual response in case of generic
				// errors. Borrowed from echo's middleware.Logger.
				c.Error(err)
			}

			log.Debug().
				Str("method", req.Method).
				Str("host", req.Host).
				Str("user_agent", req.UserAgent()).
				Dur("latency", time.Since(tn)).
				Int64("bytes_in", req.ContentLength).
				Int64("bytes_out", res.Size).
				Int("status", res.Status).
				Err(err).
				Msgf("%s %s", req.Method, req.RequestURI)

			return err
		}
	}
}
