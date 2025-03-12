// Package echopop adds middleware functions to projects using
// the Invopop API and the Echo v4 web framework.
package echopop

import (
	"net/http"
	"strings"

	"github.com/invopop/client.go/invopop"
	"github.com/labstack/echo/v4"
)

const (
	enrollmentKey      = "enrollment"
	enrollmentStateKey = "state"
	invopopClientKey   = "invopop-client"
	// HeaderEnrollmentID is the header key used to pass the enrollment ID
	HeaderEnrollmentID = "X-Enrollment-ID"
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

// AuthToken defines a middleware function that will check if the
// header contains an authentication token.
//
// If it does, the token will be included in the invopop client to be used
// to authenticate requests to the API. It is thought for endpoints where an
// oauth access token is required to access the API.
func AuthToken(ic *invopop.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tok := ""

			// extract bearer auth token
			ah := strings.Split(c.Request().Header.Get("Authorization"), "Bearer ")
			if len(ah) == 2 && ah[1] != "" {
				tok = ah[1]
			}
			if tok == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing auth token")
			}

			c.Set(invopopClientKey, ic.SetAuthToken(tok))

			return next(c)
		}
	}

}

// AuthEnrollmentByID defines a middleware function that will authenticate
// an enrollment using its ID with the Invopop API. This middleware will use
// token caching to avoid repeated authentication requests for the same enrollment.
//
// This method requires the "X-Enrollment-ID" header to be set with the enrollment ID.
// It will first try to use a cached token, and if not available or expired,
// it will request a new token and cache it.
func AuthEnrollmentByID(ic *invopop.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Get enrollment ID from header
			enrollmentID := c.Request().Header.Get("X-Enrollment-ID")
			if enrollmentID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing enrollment ID")
			}

			e, err := ic.Access().Enrollment().AuthorizeWithID(ctx, enrollmentID)
			if err != nil {
				return err
			}

			// Store the authorized client in context
			c.Set(enrollmentKey, e)
			c.Set(invopopClientKey, ic.SetAuthToken(e.Token))

			return next(c)
		}
	}
}

// GetEnrollment retrieves the enrollment object from the context.
func GetEnrollment(c echo.Context) *invopop.Enrollment {
	return c.Get(enrollmentKey).(*invopop.Enrollment)
}

// GetClient provides the Invopop client that was prepared with
// the enrollment's auth token.
func GetClient(c echo.Context) *invopop.Client {
	return c.Get(invopopClientKey).(*invopop.Client)
}
