package invopop

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadClose(t *testing.T) {
	t.Run("with a nil download", func(t *testing.T) {
		var d *Download
		assert.NoError(t, d.Close())
	})

	t.Run("with no data", func(t *testing.T) {
		d := new(Download)
		assert.NoError(t, d.Close())
	})

	t.Run("with data", func(t *testing.T) {
		d := &Download{Data: io.NopCloser(http.NoBody)}
		assert.NoError(t, d.Close())
	})
}

func TestClientDownload(t *testing.T) {
	ctx := context.Background()

	t.Run("uses the content disposition filename", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Disposition", `attachment; filename="invoice.pdf"`)
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte("%PDF-1.4"))
		}))
		defer srv.Close()

		d, err := New().Download(ctx, srv.URL+"/files/abc123")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "invoice.pdf", d.Name)
		assert.Equal(t, "application/pdf", d.Type)

		data, err := io.ReadAll(d.Data)
		require.NoError(t, err)
		assert.Equal(t, "%PDF-1.4", string(data))
	})

	t.Run("falls back to the url basename", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte("<doc/>"))
		}))
		defer srv.Close()

		d, err := New().Download(ctx, srv.URL+"/files/invoice.xml")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "invoice.xml", d.Name)
		assert.Equal(t, "application/xml", d.Type)
	})

	t.Run("falls back when the content disposition is malformed", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Disposition", "not a valid header value;;;")
			_, _ = w.Write([]byte("data"))
		}))
		defer srv.Close()

		d, err := New().Download(ctx, srv.URL+"/files/report.csv")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "report.csv", d.Name)
	})

	t.Run("falls back when the filename parameter is empty", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Disposition", "attachment")
			_, _ = w.Write([]byte("data"))
		}))
		defer srv.Close()

		d, err := New().Download(ctx, srv.URL+"/files/data.txt")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "data.txt", d.Name)
	})

	t.Run("routes spool urls through the silo service", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `file contents`)
		c := testClient(t, rec)

		d, err := c.Download(ctx, "spool:my-key")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "/silo/v1/spool/my-key", rec.last().URL.Path)
	})

	t.Run("with an invalid url", func(t *testing.T) {
		_, err := New().Download(ctx, "://not-a-url")
		assert.Error(t, err)
	})
}
