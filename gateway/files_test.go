package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
