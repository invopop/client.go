package invopop

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

func TestSiloEntriesFetchVersion(t *testing.T) {
	responder := func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/silo/v1/entries/entry-id/versions/version-id", req.URL.Path)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"version":"version-id",
				"data":{"$schema":"https://gobl.org/draft-0/envelope"}
			}`)),
		}, nil
	}
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(responder))

	out, err := c.Silo().Entries().FetchVersion(context.Background(), "entry-id", "version-id")
	require.NoError(t, err)
	assert.Equal(t, "version-id", out.Version)
	assert.JSONEq(t, `{"$schema":"https://gobl.org/draft-0/envelope"}`, string(out.Data))
}
