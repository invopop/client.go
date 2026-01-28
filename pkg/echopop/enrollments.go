package echopop

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/invopop/client.go/invopop"
	"github.com/labstack/echo/v4"
)

// Context keys
const (
	enrollmentKey      = "enrollment"
	enrollmentStateKey = "state"
	invopopClientKey   = "invopop-client"
)

// LoadEnrollment will try to load the enrollment using the request details
// and provide the enrollment and prepared client in the context.
//
// This method supports tokens provided either via the "Authorization"
// header, or a "state" query parameter, and is meant to be used
// by applications that offer a web interface via the Invopop Console.
//
// Enrollments authorized in this way will include a new token with
// additional scopes that can be used to access restricted functionality
// like updating the embedded enrollment data or accessing silo entry
// meta rows.
//
// We'd recommend using sessions instead of this method for most applications.
func LoadEnrollment(ic *invopop.Client, c echo.Context) error {
	ctx := c.Request().Context()

	// Try to read the auth token from standard locations
	tok := AuthToken(c)
	if tok == "" {
		return fmt.Errorf("%w: missing auth token", invopop.ErrAccessDenied)
	}

	// override any existing tokens in the connection
	ic = ic.SetAuthToken(tok)

	e, err := ic.Access().Enrollment().Authorize(ctx)
	if err != nil {
		if invopop.IsNotFound(err) {
			return fmt.Errorf("%w: enrollment not found", invopop.ErrAccessDenied)
		}
		return err
	}
	c.Set(enrollmentKey, e)
	c.Set(invopopClientKey, ic.SetAuthToken(e.Token))

	return nil
}

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
//
// Deprecated: You should provide your own middleware around the
// LoadEnrollment method and provide custom response handling for your
// application.
func AuthEnrollment(ic *invopop.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := LoadEnrollment(ic, c); err != nil {
				if errors.Is(err, invopop.ErrAccessDenied) {
					return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
				}
				return echo.NewHTTPError(http.StatusInternalServerError).WithInternal(err)
			}
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
