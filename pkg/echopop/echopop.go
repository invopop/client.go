// Package echopop adds middleware functions to projects using
// the Invopop API and the Echo v4 web framework.
package echopop

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

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
