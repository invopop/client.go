// Package invopop provides wrappers around the Invopop API
package invopop

import (
	"context"
	"errors"
	"fmt"
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

const (
	productionHost = "https://api.invopop.com"
)

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
func (c *Client) SetAuthToken(token string) *Client {
	c2 := *c
	c2.conn = c2.conn.SetAuthToken(token)
	c2.svc = &service{client: &c2}
	return &c2
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

// ResponseError is a wrapper around error responses from the server that will handle
// error messages.
type ResponseError struct {
	response *resty.Response

	// Code is the error code which may have been provided by the server.
	Code string `json:"code"`

	// Message contains a human readable response message from the API in the case
	// of an error.
	Message string `json:"message"`
}

// handle will wrap the resty response to provide our own Response object that
// wraps around any errors that might have happened with the connection or response.
func (r *ResponseError) handle(res *resty.Response) error {
	if res.IsSuccess() {
		return nil
	}
	r.response = res
	return r
}

// StatusCode provides the response status code, or 0 if an error occurred.
func (r *ResponseError) StatusCode() int {
	return r.response.StatusCode()
}

// Error provides the response error string.
func (r *ResponseError) Error() string {
	if r.Code != "" {
		return fmt.Sprintf("%d: (%s) %s", r.response.StatusCode(), r.Code, r.Message)
	}
	return fmt.Sprintf("%d: %v", r.response.StatusCode(), r.Message)
}

// Response provides underlying response, in case it might be useful for
// debugging.
func (r *ResponseError) Response() *resty.Response {
	return r.response
}

// IsConflict is a helper that will provide the response error object
// if the error is a conflict.
func IsConflict(err error) *ResponseError {
	return isError(err, http.StatusConflict)
}

// IsNotFound returns the error response if the status is not found.
func IsNotFound(err error) *ResponseError {
	return isError(err, http.StatusNotFound)
}

// IsForbidden returns the error response if the status is forbidden.
func IsForbidden(err error) *ResponseError {
	return isError(err, http.StatusForbidden)
}

func isError(err error, status int) *ResponseError {
	var re *ResponseError
	if errors.As(err, &re) {
		if re.StatusCode() == status {
			return re
		}
	}
	return nil
}
