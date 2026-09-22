package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// future returns a unix timestamp the given duration from now.
func future(d time.Duration) int64 {
	return time.Now().Add(d).Unix()
}

func TestSessionAuthorized(t *testing.T) {
	tests := []struct {
		name string
		sess *Session
		want bool
	}{
		{
			name: "with a valid token",
			sess: &Session{Token: testToken, TokenExpires: future(time.Hour)},
			want: true,
		},
		{
			name: "with an expired token",
			sess: &Session{Token: testToken, TokenExpires: future(-time.Hour)},
		},
		{
			name: "with no expiry",
			sess: &Session{Token: testToken},
		},
		{
			name: "with no token",
			sess: &Session{TokenExpires: future(time.Hour)},
		},
		{
			name: "empty",
			sess: new(Session),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sess.Authorized())
		})
	}
}

func TestSessionShouldRenew(t *testing.T) {
	tests := []struct {
		name string
		sess *Session
		want bool
	}{
		{
			name: "with no expiry",
			sess: new(Session),
			want: true,
		},
		{
			name: "with an expired token",
			sess: &Session{Token: testToken, TokenExpires: future(-time.Hour)},
			want: true,
		},
		{
			name: "close to expiry",
			sess: &Session{Token: testToken, TokenExpires: future(1 * time.Minute)},
			want: true,
		},
		{
			name: "well within the expiry window",
			sess: &Session{Token: testToken, TokenExpires: future(time.Hour)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sess.ShouldRenew())
		})
	}
}

func TestSessionCanStore(t *testing.T) {
	assert.False(t, new(Session).CanStore())
	assert.False(t, (&Session{EnrollmentID: testEnrollment}).CanStore())
	assert.False(t, (&Session{Token: testToken}).CanStore())
	assert.True(t, (&Session{EnrollmentID: testEnrollment, Token: testToken}).CanStore())
}

func TestSessionSetGet(t *testing.T) {
	s := new(Session)
	assert.Nil(t, s.Get("missing"), "should handle an uninitialized cache")

	s.Set("key", "value")
	assert.Equal(t, "value", s.Get("key"))
	assert.Nil(t, s.Get("other"))

	// A nil value removes the entry.
	s.Set("key", nil)
	assert.Nil(t, s.Get("key"))

	// Non-string keys and values are supported.
	type ctxKey struct{}
	s.Set(ctxKey{}, 42)
	assert.Equal(t, 42, s.Get(ctxKey{}))
}

func TestSessionSetToken(t *testing.T) {
	t.Run("without a client", func(t *testing.T) {
		s := &Session{TokenExpires: future(time.Hour)}
		s.SetToken("new-token")
		assert.Equal(t, "new-token", s.Token)
		assert.Zero(t, s.TokenExpires, "should reset expiry to force a renewal")
		assert.True(t, s.ShouldRenew())
	})

	t.Run("with a client", func(t *testing.T) {
		c := New()
		s := c.Access().NewSession()
		original := s.Client()

		s.SetToken("new-token")
		assert.Equal(t, "new-token", s.Token)
		require.NotNil(t, s.Client())
		assert.NotSame(t, original, s.Client(), "should clone the client with the token")
	})
}

func TestSessionSetOwnerID(t *testing.T) {
	s := new(Session)
	s.SetOwnerID("owner-id")
	assert.Equal(t, "owner-id", s.OwnerID)
}

func TestSessionSetFromEnrollment(t *testing.T) {
	en := &Enrollment{
		ID:           "enrollment-id",
		OwnerID:      "owner-id",
		Sandbox:      true,
		Data:         json.RawMessage(`{"key":"value"}`),
		Token:        "enrollment-token",
		TokenExpires: future(time.Hour),
	}

	t.Run("copies the enrollment details", func(t *testing.T) {
		s := new(Session)
		s.SetFromEnrollment(en)

		assert.Equal(t, "enrollment-id", s.EnrollmentID)
		assert.Equal(t, "owner-id", s.OwnerID)
		assert.True(t, s.Sandbox)
		assert.JSONEq(t, `{"key":"value"}`, string(s.Data))
		assert.Equal(t, "enrollment-token", s.Token)
		assert.Equal(t, en.TokenExpires, s.TokenExpires)
		assert.True(t, s.Authorized())
	})

	t.Run("updates the client token", func(t *testing.T) {
		c := New()
		s := c.Access().NewSession()
		original := s.Client()

		s.SetFromEnrollment(en)
		require.NotNil(t, s.Client())
		assert.NotSame(t, original, s.Client())
	})

	t.Run("leaves the client alone without a token", func(t *testing.T) {
		c := New()
		s := c.Access().NewSession()
		original := s.Client()

		s.SetFromEnrollment(&Enrollment{ID: "enrollment-id"})
		assert.Same(t, original, s.Client())
	})
}

func TestSessionUnmarshalJSON(t *testing.T) {
	t.Run("preserves the client", func(t *testing.T) {
		c := New()
		s := c.Access().NewSession()
		require.NotNil(t, s.Client())

		data := []byte(`{"eid":"enrollment-id","oid":"owner-id","sbx":true,"t":"tok","exp":1680000000}`)
		require.NoError(t, json.Unmarshal(data, s))

		assert.Equal(t, "enrollment-id", s.EnrollmentID)
		assert.Equal(t, "owner-id", s.OwnerID)
		assert.True(t, s.Sandbox)
		assert.Equal(t, testToken, s.Token)
		assert.Equal(t, int64(1680000000), s.TokenExpires)
		assert.NotNil(t, s.Client(), "client should survive unmarshaling")
	})

	t.Run("without a client", func(t *testing.T) {
		s := new(Session)
		require.NoError(t, json.Unmarshal([]byte(`{"t":"tok"}`), s))
		assert.Equal(t, testToken, s.Token)
		assert.Nil(t, s.Client())
	})

	t.Run("with invalid JSON", func(t *testing.T) {
		s := new(Session)
		assert.Error(t, json.Unmarshal([]byte(`not json`), s))
	})

	t.Run("round trips", func(t *testing.T) {
		s := &Session{
			EnrollmentID: testEnrollment,
			OwnerID:      "oid",
			Sandbox:      true,
			Meta:         map[string]string{"k": "v"},
			RedirectURI:  "https://example.com/back",
			Token:        testToken,
			TokenExpires: 1680000000,
		}
		b, err := json.Marshal(s)
		require.NoError(t, err)

		out := new(Session)
		require.NoError(t, json.Unmarshal(b, out))
		assert.Equal(t, s.EnrollmentID, out.EnrollmentID)
		assert.Equal(t, s.Meta, out.Meta)
		assert.Equal(t, s.RedirectURI, out.RedirectURI)
	})

	t.Run("does not serialize the extra cache", func(t *testing.T) {
		s := new(Session)
		s.Set("secret", "value")
		b, err := json.Marshal(s)
		require.NoError(t, err)
		assert.NotContains(t, string(b), "secret")
	})
}

func TestSessionContext(t *testing.T) {
	t.Run("stores and retrieves the session", func(t *testing.T) {
		s := &Session{EnrollmentID: testEnrollment}
		ctx := s.Context(context.Background())
		assert.Same(t, s, GetSession(ctx))
	})

	t.Run("with no session in the context", func(t *testing.T) {
		assert.Nil(t, GetSession(context.Background()))
	})
}

func TestSessionAuthorize(t *testing.T) {
	ctx := context.Background()

	t.Run("skips when the token is still fresh", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{}`)
		c := testClient(t, rec)
		s := c.Access().NewSession()
		s.Token = testToken
		s.TokenExpires = future(time.Hour)

		require.NoError(t, s.Authorize(ctx))
		assert.Empty(t, rec.requests, "no request should be made")
	})

	t.Run("without a client", func(t *testing.T) {
		s := new(Session)
		err := s.Authorize(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.EqualError(t, err, "access denied: no client available in session")
	})

	t.Run("without a token or enrollment/owner id", func(t *testing.T) {
		c := New()
		s := c.Access().NewSession()
		err := s.Authorize(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.EqualError(t, err, "access denied: no token or enrollment/owner ID provided")
	})

	t.Run("with a token", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"eid","owner_id":"oid","token":"new-tok","token_expires":9999999999}`)
		c := testOAuthClient(t, rec)
		s := c.Access().NewSession()
		s.SetToken("old-tok")

		require.NoError(t, s.Authorize(ctx))
		assert.Equal(t, testEnrollment, s.EnrollmentID)
		assert.Equal(t, "oid", s.OwnerID)
		assert.Equal(t, "new-tok", s.Token)
		assert.True(t, s.Authorized())
	})

	t.Run("with an enrollment id", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"eid","token":"new-tok","token_expires":9999999999}`)
		c := testOAuthClient(t, rec)
		s := c.Access().NewSession()
		s.EnrollmentID = testEnrollment

		require.NoError(t, s.Authorize(ctx))
		assert.Equal(t, "new-tok", s.Token)
		assert.Contains(t, rec.lastBody(), testEnrollment)
	})

	t.Run("with an owner id", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"eid","owner_id":"oid","token":"new-tok","token_expires":9999999999}`)
		c := testOAuthClient(t, rec)
		s := c.Access().NewSessionWithOwnerID("oid")

		require.NoError(t, s.Authorize(ctx))
		assert.Equal(t, "new-tok", s.Token)
		assert.Contains(t, rec.lastBody(), "oid")
	})

	t.Run("translates a not found into access denied", func(t *testing.T) {
		c := testOAuthClient(t, newRecorder(http.StatusNotFound, `{"message":"no enrollment"}`))
		s := c.Access().NewSessionWithToken(testToken)

		err := s.Authorize(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.EqualError(t, err, "access denied: application not enrolled")
	})

	t.Run("passes through other errors", func(t *testing.T) {
		c := testOAuthClient(t, newRecorder(http.StatusInternalServerError, `{"message":"boom"}`))
		s := c.Access().NewSessionWithToken(testToken)

		err := s.Authorize(ctx)
		require.Error(t, err)
		assert.False(t, errors.Is(err, ErrAccessDenied))
		assert.EqualError(t, err, "500: boom")
	})
}

func TestAccessNewSession(t *testing.T) {
	c := New()

	t.Run("new session", func(t *testing.T) {
		s := c.Access().NewSession()
		require.NotNil(t, s)
		assert.Same(t, c, s.Client())
		assert.Empty(t, s.Token)
	})

	t.Run("new session with token", func(t *testing.T) {
		s := c.Access().NewSessionWithToken(testToken)
		require.NotNil(t, s)
		assert.Equal(t, testToken, s.Token)
		assert.NotNil(t, s.Client())
	})

	t.Run("new session with owner id", func(t *testing.T) {
		s := c.Access().NewSessionWithOwnerID("oid")
		require.NotNil(t, s)
		assert.Equal(t, "oid", s.OwnerID)
	})
}
