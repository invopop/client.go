package invopop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	sessionTokenTTL                   = 300 // 5 minutes
	sessionKey      invopopContextKey = "session"
)

// Session provides an opinionated session management structure for usage alongside
// application enrollments inside Invopop. Content is based on the enrollment provided
// by the Access service. Sessions may be persisted to secure temporary stores, such as
// cookies or key-value databases.
//
// Sessions should be used within your business or domain logic and can be used as a
// replacement of a regular Client instance.
//
// Instantiate sessions using the Access NewSession or related methods, even if you intend to
// unmarshal them from a stored source to ensure that there is always a client available.
//
// Sessions must be authorized before use by calling the Authorize method, which will
// ensure that the token is valid and renewed if necessary. If no token is available in
// the session, but an Enrollment ID or Owner ID is provided, those will be used to
// authorize the session instead.
//
// Important: sessions should only be stored in cookies if the application using them
// will be used independently. For embedded applications, such as those running inside
// the Invopop Console, sessions should be created on each request.
type Session struct {
	// EnrollmentID is the unique ID of the enrollment.
	EnrollmentID string `json:"eid,omitempty"`
	// OwnerID is the unique ID of the entity this enrollment belongs to.
	OwnerID string `json:"oid,omitempty"`
	// Sandbox indicates whether this enrollment is for a sandbox or live workspace.
	Sandbox bool `json:"sbx,omitempty"`

	// Data contains any config data stored inside the enrollment. This should not be
	// modified.
	Data json.RawMessage `json:"data,omitempty"`

	// Meta contains any metadata stored alongside the base enrollment
	// details that might be useful for the app. Use the Set and Get methods to manage
	// this map. Only string key-value pairs are supported for simplicity in
	// serialization.
	Meta map[string]string `json:"meta,omitempty"`

	// RedirectURI is a convenience field to store a redirection URI that may be
	// used as part of an authentication flow to redirect the user back to where
	// the came from.
	RedirectURI string `json:"redirect_uri,omitempty"`

	// Token is the authentication token provided by the enrollment, alongside
	// the expiration unix timestamp.
	Token        string `json:"t,omitempty"`
	TokenExpires int64  `json:"exp,omitempty"`

	// extra is used to temporarily store any extra values in the session that may
	// be used in the same request and must not be serialized.
	extra map[any]any `json:"-"`

	// client contains the invopop client prepared with the session's token, when loaded.
	client *Client
}

// Authorize will try to authorize the session by confirming that
// the token is valid with the Access service, renewing in the process.
//
// Before making unnecessary requests, a check is made to see if the currently
// provided token is still valid based on the expiration timestamp. To force
// renewal, set the TokenExpires field to zero.
//
// If an Enrollment or Owner ID is provided inside the session and no token is
// present, those will be used to try to authorize the session instead.
//
// If no client is available, or no token is present, an error will be returned.
func (s *Session) Authorize(ctx context.Context) error {
	if !s.ShouldRenew() {
		return nil
	}
	if s.client == nil {
		return fmt.Errorf("%w: no client available in session", ErrAccessDenied)
	}
	var en *Enrollment
	var err error
	if s.Token != "" {
		en, err = s.client.Access().Enrollment().Authorize(ctx)
	} else if s.EnrollmentID != "" {
		en, err = s.client.Access().Enrollment().AuthorizeWithID(ctx, s.EnrollmentID)
	} else if s.OwnerID != "" {
		en, err = s.client.Access().Enrollment().AuthorizeWithOwnerID(ctx, s.OwnerID)
	} else {
		return fmt.Errorf("%w: no token or enrollment/owner ID provided", ErrAccessDenied)
	}
	if err != nil {
		if IsNotFound(err) {
			return fmt.Errorf("%w: application not enrolled", ErrAccessDenied)
		}
		return err
	}
	s.SetFromEnrollment(en)
	return nil
}

// Client provides the invopop client prepared with the session's token, which may be nil
// if the session was not initialized with a client or token.
func (s *Session) Client() *Client {
	return s.client
}

// SetFromEnrollment will update the session details based on the provided
// enrollment object.
func (s *Session) SetFromEnrollment(e *Enrollment) {
	s.EnrollmentID = e.ID
	s.OwnerID = e.OwnerID
	s.Sandbox = e.Sandbox
	s.Data = e.Data
	s.Token = e.Token
	s.TokenExpires = e.TokenExpires
	if s.client != nil && e.Token != "" {
		s.client = s.client.SetAuthToken(e.Token)
	}
}

// Set will store a key-value pair inside the session much like in a context object
// that can be later retrieved using the Get method. This is useful to caching details
// such as a database connection or similar complex object.
func (s *Session) Set(key, value any) {
	if s.extra == nil {
		s.extra = make(map[any]any)
	}
	if value == nil {
		delete(s.extra, key)
		return
	}
	s.extra[key] = value
}

// Get will retrieve a value from the session's cache or return nil.
func (s *Session) Get(key any) any {
	if s.extra == nil {
		return nil
	}
	return s.extra[key]
}

// SetToken will update the session's token value and prepare for an authorization
// request to be sent to the Invopop API.
func (s *Session) SetToken(tok string) {
	s.Token = tok
	s.TokenExpires = 0
	if s.client != nil {
		s.client = s.client.WithAuthToken(tok)
	}
}

// SetOwnerID will update the session's owner ID value in preparation for an authorization
// request to be sent to the Invopop API.
func (s *Session) SetOwnerID(oid string) {
	s.OwnerID = oid
}

// Authorized indicates whether the session has a valid token that is not expired.
// The server may have a different state, so this does not guarantee that requests
// will succeed, but is a good pre-check before making calls.
func (s *Session) Authorized() bool {
	return s.Token != "" && s.TokenExpires != 0 && time.Now().Unix() < s.TokenExpires
}

// ShouldRenew indicates whether the session's token is close to expiration
// and should be renewed.
func (s *Session) ShouldRenew() bool {
	if s.TokenExpires == 0 {
		return true
	}
	tn := time.Now().Unix()
	// renew if less than 5 minutes to expiration
	return (s.TokenExpires - tn) < sessionTokenTTL
}

// CanStore indicates whether the session has sufficient data to be stored.
func (s *Session) CanStore() bool {
	return s.EnrollmentID != "" && s.Token != ""
}

// UnmarshalJSON implements the json.Unmarshaler interface to
// ensure that any token in the payload will be added to the client
// automatically.
func (s *Session) UnmarshalJSON(data []byte) error {
	type sessionAlias Session
	var sa sessionAlias
	if err := json.Unmarshal(data, &sa); err != nil {
		return err
	}
	sa.client = s.client // ensure client preserved
	*s = Session(sa)
	if s.Token != "" && s.client != nil {
		s.client = s.client.SetAuthToken(s.Token)
	}
	return nil
}

// Context adds the session to the context so that it can be easily re-used inside
// other parts of the application. Use this sparingly, ideally you want to be passing
// the session directly between method calls, but given that a session may have different
// credentials for each incoming request, the context can be a lot more convenient.
func (s *Session) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

// GetSession tries to extract a session object from the provided context.
func GetSession(ctx context.Context) *Session {
	s, ok := ctx.Value(sessionKey).(*Session)
	if !ok {
		return nil
	}
	return s
}
