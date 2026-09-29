package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareCreateFileFromData(t *testing.T) {
	t.Run("detects the mime type of the data", func(t *testing.T) {
		gw := new(Client)
		req := new(CreateFile)
		data := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")

		gw.prepareCreateFileFromData(req, data)

		assert.Equal(t, "text/xml; charset=utf-8", req.Mime)
	})

	t.Run("doesn't overwrite the mime type if already set", func(t *testing.T) {
		gw := new(Client)
		req := new(CreateFile)
		req.Mime = "text/plain"
		data := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")

		gw.prepareCreateFileFromData(req, data)

		assert.Equal(t, "text/plain", req.Mime)
	})
}

func TestPrepareCreateFileFromDataAttributes(t *testing.T) {
	gw := new(Client)
	req := new(CreateFile)
	data := []byte("hello world")

	gw.prepareCreateFileFromData(req, data)

	// SHA256 of "hello world"
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", req.Sha256)
	assert.Equal(t, int32(11), req.Size)
	assert.Equal(t, "text/plain; charset=utf-8", req.Mime)

	t.Run("with empty data", func(t *testing.T) {
		gw := new(Client)
		req := new(CreateFile)
		gw.prepareCreateFileFromData(req, []byte{})

		// SHA256 of an empty payload.
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", req.Sha256)
		assert.Zero(t, req.Size)
	})
}

func TestFileUploadURL(t *testing.T) {
	f := &File{
		Id:          "file-id",
		SiloEntryId: "entry-id",
		Name:        "invoice.pdf",
		Hash:        "abc123",
	}

	t.Run("without a public base url", func(t *testing.T) {
		gw := new(Client)
		url, err := gw.fileUploadURL(f)
		assert.EqualError(t, err, "missing silo public base url")
		assert.Empty(t, url)
	})

	t.Run("with a public base url", func(t *testing.T) {
		gw := new(Client)
		gw.siloPublicBaseURL = "https://silo.example.com"
		url, err := gw.fileUploadURL(f)
		require.NoError(t, err)
		assert.Equal(t, "https://silo.example.com/entry-id/file-id/invoice.pdf?h=abc123", url)
	})
}
