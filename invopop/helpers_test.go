package invopop

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

// recorder captures the requests made through a test client and replies with a
// canned response, so that request paths, methods and bodies can be asserted
// without touching the network.
type recorder struct {
	t *testing.T

	// status and body define the canned response. A zero status means 200.
	status int
	body   string

	// requests holds every request seen, in order.
	requests []*http.Request
	// bodies holds the raw request payloads, in order.
	bodies []string
}

func (r *recorder) respond(req *http.Request) (*http.Response, error) {
	r.t.Helper()

	payload := ""
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			r.t.Fatalf("reading request body: %v", err)
		}
		payload = string(b)
	}
	r.requests = append(r.requests, req)
	r.bodies = append(r.bodies, payload)

	status := r.status
	if status == 0 {
		status = http.StatusOK
	}
	body := r.body
	if body == "" {
		body = "{}"
	}

	resp := jsonResponse(body)
	resp.StatusCode = status
	resp.Request = req
	return resp, nil
}

// jsonResponse builds a canned JSON response for a stubbed round tripper.
func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// last returns the most recent request, failing the test if none were made.
func (r *recorder) last() *http.Request {
	r.t.Helper()
	if len(r.requests) == 0 {
		r.t.Fatal("no requests recorded")
	}
	return r.requests[len(r.requests)-1]
}

// lastBody returns the payload of the most recent request.
func (r *recorder) lastBody() string {
	r.t.Helper()
	if len(r.bodies) == 0 {
		r.t.Fatal("no requests recorded")
	}
	return r.bodies[len(r.bodies)-1]
}

// testClient prepares a client whose connection is backed by the recorder.
func testClient(t *testing.T, rec *recorder) *Client {
	t.Helper()
	rec.t = t
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(rec.respond)).
		SetBaseURL("https://api.test")
	return c
}

// newRecorder builds a recorder returning the provided status and body.
func newRecorder(status int, body string) *recorder {
	return &recorder{status: status, body: body}
}

// testOAuthClient prepares a recorder-backed client that also carries the OAuth
// app credentials required by the enrollment authorization endpoints.
func testOAuthClient(t *testing.T, rec *recorder) *Client {
	t.Helper()
	c := testClient(t, rec)
	c.clientID = "client-id"
	c.clientSecret = "client-secret"
	return c
}

// Shared fixture values, kept as constants so that the repeated literals stay
// consistent across the package's tests.
const (
	testEntryID     = "entry-1"
	testToken       = "tok"
	testEnrollment  = "eid"
	testCursor      = "abc"
	testCreatedAt   = "2023-08-02T00:00:00.000Z"
	testCompletedAt = "2024-11-01T00:00:00.000Z"
	testStepID      = "step-1"
	testFolder      = "sales"
	testFileName    = "f.pdf"
	testSchema      = "bill/invoice"

	errMissingKey  = "missing key"
	errMissingData = "missing data"
)
