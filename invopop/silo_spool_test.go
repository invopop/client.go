package invopop

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiloSpoolValidation(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{}`)
	svc := testClient(t, rec).Silo().Spool()

	tests := []struct {
		name string
		call func() error
		err  string
	}{
		{
			name: "upload without a name",
			call: func() error { _, err := svc.Upload(ctx, "", "text/plain", []byte("x")); return err },
			err:  "missing name",
		},
		{
			name: "upload without a media type",
			call: func() error { _, err := svc.Upload(ctx, "f.txt", "", []byte("x")); return err },
			err:  "missing media type",
		},
		{
			name: "upload without data",
			call: func() error { _, err := svc.Upload(ctx, "f.txt", "text/plain", nil); return err },
			err:  errMissingData,
		},
		{
			name: "upload with empty data",
			call: func() error { _, err := svc.Upload(ctx, "f.txt", "text/plain", []byte{}); return err },
			err:  errMissingData,
		},
		{
			name: "download without a key",
			call: func() error { _, err := svc.Download(ctx, ""); return err },
			err:  errMissingKey,
		},
		{
			name: "delete without a key",
			call: func() error { return svc.Delete(ctx, "") },
			err:  errMissingKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.call(), tt.err)
		})
	}

	assert.Empty(t, rec.requests, "validation should fail before any request is made")
}

func TestSiloSpoolUpload(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the storage key", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"key":"spool-key"}`)
		c := testClient(t, rec)

		key, err := c.Silo().Spool().Upload(ctx, "invoice.pdf", "application/pdf", []byte("data"))
		require.NoError(t, err)
		assert.Equal(t, "spool-key", key)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/silo/v1/spool", req.URL.Path)

		body := rec.lastBody()
		assert.Contains(t, body, `"name":"invoice.pdf"`)
		assert.Contains(t, body, `"type":"application/pdf"`)
		assert.Contains(t, body, `"data":"ZGF0YQ=="`, "data should be base64 encoded")
	})

	t.Run("with an error response", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusForbidden, `{"message":"not enrolled"}`))

		key, err := c.Silo().Spool().Upload(ctx, "f.txt", "text/plain", []byte("x"))
		require.Error(t, err)
		assert.True(t, IsForbidden(err))
		assert.Empty(t, key)
	})
}

func TestSiloSpoolDownload(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the file contents", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `file contents`)
		c := testClient(t, rec)

		d, err := c.Silo().Spool().Download(ctx, "spool-key")
		require.NoError(t, err)
		defer d.Close() //nolint:errcheck

		assert.Equal(t, "/silo/v1/spool/spool-key", rec.last().URL.Path)
		assert.Equal(t, "spool-key", d.Name, "should fall back to the path basename")

		data, err := io.ReadAll(d.Data)
		require.NoError(t, err)
		assert.Equal(t, "file contents", string(data))
	})

	t.Run("with an error response", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusNotFound, `{"message":"gone"}`))

		d, err := c.Silo().Spool().Download(ctx, "spool-key")
		require.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, d)
	})
}

func TestSiloSpoolDelete(t *testing.T) {
	ctx := context.Background()

	rec := newRecorder(http.StatusOK, `{}`)
	c := testClient(t, rec)

	require.NoError(t, c.Silo().Spool().Delete(ctx, "spool-key"))

	req := rec.last()
	assert.Equal(t, http.MethodDelete, req.Method)
	assert.Equal(t, "/silo/v1/spool/spool-key", req.URL.Path)
}
