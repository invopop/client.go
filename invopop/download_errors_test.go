package invopop

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

// closeTracker records whether a response body was closed.
type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

// errorClient replies to every request with the error status and body,
// keeping the body so the test can check it was closed.
func errorClient(t *testing.T, status int, ctype, body string) (*Client, *closeTracker) {
	t.Helper()
	tracker := &closeTracker{Reader: strings.NewReader(body)}
	respond := func(req *http.Request) (*http.Response, error) {
		h := make(http.Header)
		h.Set("Content-Type", ctype)
		return &http.Response{StatusCode: status, Header: h, Body: tracker, Request: req}, nil
	}
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(respond)).SetBaseURL("https://api.test")
	return c, tracker
}

func TestDownloadErrors(t *testing.T) {
	ctx := context.Background()
	downloads := map[string]func(*Client) (any, error){
		"file": func(c *Client) (any, error) {
			return c.Silo().Files().Download(ctx, testEntryID, testFileID)
		},
		"spool": func(c *Client) (any, error) {
			return c.Silo().Spool().Download(ctx, "spool-key")
		},
		"entry export": func(c *Client) (any, error) {
			return c.Silo().Entries().Export(ctx, &ExportSiloEntry{ID: testEntryID, Formats: []string{testFormatCII}})
		},
	}

	for name, download := range downloads {
		t.Run(name+" reads the API's message", func(t *testing.T) {
			c, body := errorClient(t, http.StatusNotFound, "application/json", `{"code":"not_found","message":"entry not found"}`)
			out, err := download(c)
			require.Error(t, err)
			assert.Nil(t, out)
			assert.True(t, IsNotFound(err))
			re := AsResponseError(err)
			require.NotNil(t, re)
			assert.Equal(t, "not_found", re.Code)
			assert.Equal(t, "entry not found", re.Message)
			assert.True(t, body.closed, "the body must be closed when nothing is returned to close it")
		})

		t.Run(name+" falls back to the status text", func(t *testing.T) {
			c, body := errorClient(t, http.StatusBadGateway, "text/html", "<html>Bad Gateway</html>")
			out, err := download(c)
			require.Error(t, err)
			assert.Nil(t, out)
			assert.EqualError(t, err, "502: Bad Gateway")
			assert.True(t, body.closed)
		})
	}
}

func TestClientDownloadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html>Not Found</html>"))
	}))
	defer srv.Close()

	d, err := New().Download(context.Background(), srv.URL+"/files/invoice.pdf")
	require.Error(t, err, "an error page must not pass for the file")
	assert.Nil(t, d)
	assert.True(t, IsNotFound(err))
	assert.EqualError(t, err, "404: Not Found")
}
