// Package invopop provides wrappers around the Invopop API
package invopop

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Client manages communication with the Invopop API.
type Client struct {
	conn *resty.Client

	// OAuth credentials which may have been configured	and
	// will only be used if needed by specific endpoints.
	clientID     string
	clientSecret string

	svc *service
}

type invopopClientKey string

const (
	productionHost                  = "https://api.invopop.com"
	clientKey      invopopClientKey = "invopop-client"
)

// HTTPClient returns the underlying HTTP client (useful for mocking).
func (c *Client) HTTPClient() *http.Client {
	return c.conn.GetClient()
}

// Utils provides access to the utils service.
func (c *Client) Utils() *UtilsService {
	return (*UtilsService)(c.svc)
}

// Sequence provides access to the sequence service.
func (c *Client) Sequence() *SequenceService {
	return (*SequenceService)(c.svc)
}

// Transform provides access to the transform service.
func (c *Client) Transform() *TransformService {
	return (*TransformService)(c.svc)
}

// Silo provides access to the silo service.
func (c *Client) Silo() *SiloService {
	return (*SiloService)(c.svc)
}

// Access provides a client to the access resources.
func (c *Client) Access() *AccessService {
	return (*AccessService)(c.svc)
}

// ClientOption defines options when initializing a client.
type ClientOption func(c *Client)

type service struct {
	client *Client
}

// New returns a new Invopop API client.
func New(opts ...ClientOption) *Client {
	c := new(Client)

	c.conn = resty.New().
		SetBaseURL(productionHost)

	for _, opt := range opts {
		opt(c)
	}

	// Reuse a single struct instead of allocating one for each service on the heap.
	c.svc = new(service)
	c.svc.client = c

	return c
}

// WithConfig will use the provided configuration to set up the client.
func WithConfig(conf *Config) ClientOption {
	return func(c *Client) {
		if conf.BaseURL != "" {
			c.conn = c.conn.SetBaseURL(conf.BaseURL)
		}
		if conf.ClientID != "" {
			c.clientID = conf.ClientID
		}
		if conf.ClientSecret != "" {
			c.clientSecret = conf.ClientSecret
		}
	}
}

// WithAuthToken can be used to set the authentication token
// for the API Key to use with the API.
func WithAuthToken(token string) ClientOption {
	return func(c *Client) {
		c.conn = c.conn.SetAuthToken(token)
	}
}

// WithOAuthClient can be used to configure the OAuth client ID and secret
// which is useful for applications registered with Invopop.
func WithOAuthClient(id, secret string) ClientOption {
	return func(c *Client) {
		c.clientID = id
		c.clientSecret = secret
	}
}

// SetAuthToken will set the authentication token inside a
// new client instance. This is useful for dealing with multiple
// connections that don't necessarily share the same token, such
// as when building apps that use enrollments to authenticate
// sessions.
//
// Note: This doesn't interact with the token cache. Use CacheEnrollmentToken
// and GetCachedEnrollmentToken for token caching.
func (c *Client) SetAuthToken(token string) *Client {
	c2 := *c
	c2.conn = c2.conn.Clone().SetAuthToken(token)
	c2.svc = &service{client: &c2}
	return &c2
}

// Context adds the current client model to the context so that it can be
// easily re-used inside other parts of the application. Use this sparingly,
// ideally you want to be passing the client directly, but given that a client
// may have an auth token for each connection, using the context can be
// a lot more convenient.
func (c *Client) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// GetClient tries to extract a client object from the context.
func GetClient(ctx context.Context) *Client {
	c, ok := ctx.Value(clientKey).(*Client)
	if !ok {
		return nil
	}
	return c
}

func (c *Client) get(ctx context.Context, path string, body interface{}) error {
	re := new(ResponseError)
	res, err := c.conn.R().
		SetContext(ctx).
		SetResult(body).
		SetError(re).
		Get(path)
	if err != nil {
		return err
	}
	return re.handle(res)
}

func (c *Client) post(ctx context.Context, path string, in, out any) error {
	re := new(ResponseError)
	res, err := c.conn.R().
		SetContext(ctx).
		SetBody(in).
		SetError(re).
		SetResult(out).
		Post(path)
	if err != nil {
		return err
	}
	return re.handle(res)
}

func (c *Client) put(ctx context.Context, path string, in, out any) error {
	re := new(ResponseError)
	res, err := c.conn.R().
		SetContext(ctx).
		SetBody(in).
		SetError(re).
		SetResult(out).
		Put(path)
	if err != nil {
		return err
	}
	return re.handle(res)
}

func (c *Client) patch(ctx context.Context, path string, in, out any) error {
	re := new(ResponseError)
	res, err := c.conn.R().
		SetContext(ctx).
		SetBody(in).
		SetError(re).
		SetResult(out).
		Patch(path)
	if err != nil {
		return err
	}
	return re.handle(res)
}
