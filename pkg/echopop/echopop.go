// Package echopop adds middleware functions to projects using
// the Invopop API and the Echo v4 web framework.
package echopop

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/invopop/client.go/invopop"
	"github.com/labstack/echo/v4"
)

// Context keys
const (
	enrollmentKey      = "enrollment"
	enrollmentStateKey = "state"
	invopopClientKey   = "invopop-client"
)

// AuthEnrollment defines a middleware function that will authenticate
// an enrollment with the Invopop API. This middleware will only
// work if the invopop client has been prepared using the OAuth Client
// ID and Secret.
//
// This method supports tokens provided either via the "Authorization"
// header, or a "state" query parameter, and is meant to be used
// by applications that offer a web interface via the Invopop Console.
//
// Enrollments authorized in this way will include a new token with
// additional scopes that can be used to access restricted functionality
// like updating the embedded enrollment data or accessing silo entry
// meta rows.
func AuthEnrollment(ic *invopop.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			tok := ""

			// extract bearer auth token
			ah := strings.Split(c.Request().Header.Get("Authorization"), "Bearer ")
			if len(ah) == 2 && ah[1] != "" {
				tok = ah[1]
			}
			if tok == "" {
				// try to use OAuth 2.0 state query param
				tok = c.QueryParam(enrollmentStateKey)
			}
			if tok == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing auth token")
			}

			// override any existing tokens in the connection
			ic = ic.SetAuthToken(tok)

			e, err := ic.Access().Enrollment().Authorize(ctx)
			if err != nil {
				return err
			}
			c.Set(enrollmentKey, e)
			c.Set(invopopClientKey, ic.SetAuthToken(e.Token))

			return next(c)
		}
	}

}

// GetEnrollment retrieves the enrollment object from the context.
func GetEnrollment(c echo.Context) *invopop.Enrollment {
	if en, ok := c.Get(enrollmentKey).(*invopop.Enrollment); ok {
		return en
	}
	return nil
}

// GetClient provides the Invopop client that was prepared with
// the enrollment's auth token.
func GetClient(c echo.Context) *invopop.Client {
	if c, ok := c.Get(invopopClientKey).(*invopop.Client); ok {
		return c
	}
	return nil
}

// Render will render the provided Templ Component.
//
// Usage example:
//
//	func (ct *controller) config(c echo.Context) error {
//		return echopop.Render(c, http.StatusOK, app.Configure())
//	}
func Render(c echo.Context, status int, t templ.Component) error {
	c.Response().Writer.WriteHeader(status)

	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
