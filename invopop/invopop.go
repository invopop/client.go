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

	utils     *UtilsService
	sequence  *SequenceService
	transform *TransformService
	silo      *SiloService
	access    *AccessService
}

const (
	productionHost = "https://api.invopop.com"
)

// Utils provides access to the utils service.
func (c *Client) Utils() *UtilsService {
	return c.utils
}

// Sequence provides access to the sequence service.
func (c *Client) Sequence() *SequenceService {
	return c.sequence
}

// Transform provides access to the transform service.
func (c *Client) Transform() *TransformService {
	return c.transform
}

// Silo provides access to the silo service.
func (c *Client) Silo() *SiloService {
	return c.silo
}

// Access provides a client to the access resources.
func (c *Client) Access() *AccessService {
	return c.access
}

type requestOptions struct {
	wait int
}

// RequestOption is used to define options for the request.
type RequestOption func(o *requestOptions)

type service struct {
	client *Client
}

// New returns a new Invopop API client.
func New() *Client {
	c := new(Client)

	c.conn = resty.New().
		SetBaseURL(productionHost)

	// Reuse a single struct instead of allocating one for each service on the heap.
	common := new(service)
	common.client = c

	c.utils = (*UtilsService)(common)
	c.sequence = (*SequenceService)(common)
	c.transform = (*TransformService)(common)
	c.silo = (*SiloService)(common)
	c.access = newAccessService(common)

	return c
}

// SetBaseURL can be used to override the default base URL for the client. This is
// useful for testing the client against a local or staging server.
func (c *Client) SetBaseURL(url string) {
	c.conn = c.conn.SetBaseURL(url)
}

// SetAuthToken updates the client connection's auth token.
func (c *Client) SetAuthToken(token string) {
	c.conn = c.conn.SetAuthToken(token)
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
		Put(path)
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

// WithWait adds a wait parameter to the query where it is supported. Typically
// this is used with job requests that may take longer to respond.
func WithWait(t int) RequestOption {
	return func(o *requestOptions) {
		o.wait = t
	}
}

func handleOptions(opts []RequestOption) *requestOptions {
	ro := new(requestOptions)
	for _, o := range opts {
		o(ro)
	}
	return ro
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
