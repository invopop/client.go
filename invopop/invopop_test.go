package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

func Test_put(t *testing.T) {
	tests := []struct {
		name      string
		responder testy.HTTPResponder
		path      string
		body      interface{}
		err       string
	}{
		{
			name: "unmarshalable body",
			body: json.RawMessage("this is not JSON"),
			err:  `json: error calling MarshalJSON for type json.RawMessage: invalid character 'h' in literal true (expecting 'r')`,
		},
		{
			name: "network error",
			responder: func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
			body: map[string]string{"foo": "bar"},
			err:  `Put "": network error`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				conn: resty.NewWithClient(testy.HTTPClient(tt.responder)),
			}
			err := c.put(context.Background(), tt.path, tt.body, tt.body)
			assert.EqualError(t, err, tt.err)
		})
	}
}

func TestHTTPClient(t *testing.T) {
	hc := new(http.Client)
	c := &Client{
		conn: resty.NewWithClient(hc),
	}
	assert.Equal(t, hc, c.HTTPClient())
}

func TestNew(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		c := New()
		require.NotNil(t, c)
		assert.Equal(t, productionHost, c.conn.BaseURL())
		assert.Empty(t, c.clientID)
		assert.Empty(t, c.clientSecret)
	})

	t.Run("with an auth token", func(t *testing.T) {
		c := New(WithAuthToken("my-token"))
		require.NotNil(t, c)
		assert.Equal(t, "my-token", c.conn.AuthToken())
	})

	t.Run("with an oauth client", func(t *testing.T) {
		c := New(WithOAuthClient("id", "secret"))
		assert.Equal(t, "id", c.clientID)
		assert.Equal(t, "secret", c.clientSecret)
	})

	t.Run("with a config", func(t *testing.T) {
		c := New(WithConfig(&Config{
			BaseURL:      "https://api.test.invopop.com",
			ClientID:     "id",
			ClientSecret: "secret",
		}))
		assert.Equal(t, "https://api.test.invopop.com", c.conn.BaseURL())
		assert.Equal(t, "id", c.clientID)
		assert.Equal(t, "secret", c.clientSecret)
	})

	t.Run("with an empty config", func(t *testing.T) {
		c := New(WithConfig(new(Config)))
		assert.Equal(t, productionHost, c.conn.BaseURL(), "should keep the default host")
		assert.Empty(t, c.clientID)
	})

	t.Run("with a config missing the secret", func(t *testing.T) {
		c := New(WithConfig(&Config{ClientID: "id"}))
		assert.Empty(t, c.clientID, "both credentials are required")
	})
}

func TestClientServices(t *testing.T) {
	c := New()

	assert.NotNil(t, c.Utils())
	assert.NotNil(t, c.Sequence())
	assert.NotNil(t, c.Transform())
	assert.NotNil(t, c.Silo())
	assert.NotNil(t, c.Access())

	// All services share the client's single service struct.
	assert.Same(t, c, c.Silo().Entries().client)
	assert.Same(t, c, c.Access().Enrollment().client)
}

func TestClientWithAuthToken(t *testing.T) {
	c := New(WithAuthToken("original"))
	c2 := c.WithAuthToken("replacement")

	require.NotSame(t, c, c2, "should return a new client")
	assert.Equal(t, "original", c.conn.AuthToken(), "the original should be untouched")
	assert.Equal(t, "replacement", c2.conn.AuthToken())
	assert.Same(t, c2, c2.Silo().Entries().client, "services should point at the new client")
}

func TestClientSetAuthToken(t *testing.T) {
	// Deprecated alias of WithAuthToken.
	c := New(WithAuthToken("original"))
	c2 := c.SetAuthToken("replacement")

	require.NotSame(t, c, c2)
	assert.Equal(t, "replacement", c2.conn.AuthToken())
	assert.Equal(t, "original", c.conn.AuthToken())
}

func TestClientContext(t *testing.T) {
	t.Run("stores and retrieves the client", func(t *testing.T) {
		c := New()
		ctx := c.Context(context.Background())
		assert.Same(t, c, GetClient(ctx))
	})

	t.Run("with no client in the context", func(t *testing.T) {
		assert.Nil(t, GetClient(context.Background()))
	})
}

func TestUtilsPing(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"ping":"pong"}`)
	c := testClient(t, rec)

	p := new(Ping)
	require.NoError(t, c.Utils().Ping(context.Background(), p))
	assert.Equal(t, "pong", p.Ping)
	assert.Equal(t, "/utils/v1/ping", rec.last().URL.Path)
}

func TestClientVerbs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		call   func(c *Client) error
		method string
		body   bool
	}{
		{
			name:   "get",
			call:   func(c *Client) error { return c.get(ctx, "/test", new(Ping)) },
			method: http.MethodGet,
		},
		{
			name:   "post",
			call:   func(c *Client) error { return c.post(ctx, "/test", &Ping{Ping: "in"}, new(Ping)) },
			method: http.MethodPost,
			body:   true,
		},
		{
			name:   "put",
			call:   func(c *Client) error { return c.put(ctx, "/test", &Ping{Ping: "in"}, new(Ping)) },
			method: http.MethodPut,
			body:   true,
		},
		{
			name:   "patch",
			call:   func(c *Client) error { return c.patch(ctx, "/test", &Ping{Ping: "in"}, new(Ping)) },
			method: http.MethodPatch,
			body:   true,
		},
		{
			name:   "delete",
			call:   func(c *Client) error { return c.delete(ctx, "/test", new(Ping)) },
			method: http.MethodDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(http.StatusOK, `{"ping":"pong"}`)
			c := testClient(t, rec)

			require.NoError(t, tt.call(c))
			req := rec.last()
			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, "/test", req.URL.Path)
			if tt.body {
				assert.Contains(t, rec.lastBody(), `"ping":"in"`)
			}
		})

		t.Run(tt.name+" with an error response", func(t *testing.T) {
			c := testClient(t, newRecorder(http.StatusNotFound, `{"message":"nope"}`))
			err := tt.call(c)
			require.Error(t, err)
			assert.True(t, IsNotFound(err))
		})
	}
}
