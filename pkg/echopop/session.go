package echopop

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/invopop/client.go/invopop"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	sessionCookieName = "_echopop"
	sessionCtxKey     = "session"
)

// LoadSession provides the session middleware for usage in routes that may
// use a session. This will only try and prepare the session object based
// on the presence of the `Authorization` header, `state` query parameter,
// or a cookie, and will not enforce that a session is authorized.
//
// Depending on your use-case, you may to follow this middleware up with an
// Authorize call on the session to ensure that it is valid and populate it with
// data from the enrollment.
//
// While storage in cookies is permitted, as a rule this is not recommended,
// and especially not for embedded applications running inside the Invopop Console
// where sessions stored in cookies would be shared between multiple browser tabs
// potentially showing different workspaces.
func LoadSession(ic *invopop.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess := ic.Access().NewSession()

			// Try to extract a token from the headers
			if tok := AuthToken(c); tok != "" {
				sess.SetToken(tok)
			} else {
				// Try to extract the session from the cookie
				err := extractSessionFromCookie(c, sess)
				if err != nil {
					return fmt.Errorf("extracting session from cookie: %w", err)
				}
			}

			// Prepare the session for use in the rest of the context.
			c.Set(sessionCtxKey, sess)

			return next(c)
		}
	}
}

func extractSessionFromCookie(c echo.Context, sess *invopop.Session) error {
	ck, err := session.Get(sessionCookieName, c)
	if err != nil {
		return fmt.Errorf("from cookie: %w", err)
	}

	// If there is something we can parse, try to use it. This implies
	// that a previous session was authenticated and stored.
	if val, ok := ck.Values[sessionCtxKey]; ok {
		err = json.Unmarshal([]byte(val.(string)), sess)
		if err != nil {
			return fmt.Errorf("unmarshaling session: %w", err)
		}
	}

	return nil
}

// GetSession will retrieve the session object from the context assuming that it was already
// prepared using the echopop Service's LoadSession middleware.
func GetSession(c echo.Context) *invopop.Session {
	sess, ok := c.Get(sessionCtxKey).(*invopop.Session)
	if !ok {
		return nil
	}
	return sess
}

// StoreSessionCookie will store the session object into a secure cookie in the response headers.
// Only use cookies when the Authorization cannot be used.
func StoreSessionCookie(c echo.Context, sess *invopop.Session) error {
	cs, err := session.Get(sessionCookieName, c)
	if err != nil {
		// ignore session errors
		return fmt.Errorf("preparing session: %w", err)
	}

	tn := time.Now().Unix()
	if sess.TokenExpires > 0 && tn >= sess.TokenExpires {
		// session expired, clear it
		cs.Options = &sessions.Options{
			MaxAge: -1,
		}
	} else {
		cs.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   int(sess.TokenExpires - tn),
			HttpOnly: false,
		}
		b, err := json.Marshal(sess)
		if err != nil {
			return fmt.Errorf("marshaling session: %w", err)
		}
		cs.Values[sessionCtxKey] = string(b)
	}

	return cs.Save(c.Request(), c.Response())
}
