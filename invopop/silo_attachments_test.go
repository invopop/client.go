package invopop

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

const (
	testFileID   = "file-id"
	testEntryID  = "entry-id"
	testFileName = "invoice.xml"
	testFileMIME = "application/xml"
	testFilePath = "/silo/v1/entries/" + testEntryID + "/files/" + testFileID
	testDataPath = testFilePath + "/data"
)

func fileResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestSiloFilesCreateWithoutData(t *testing.T) {
	t.Run("requires the details the silo needs to expect an upload", func(t *testing.T) {
		c := New()
		_, err := c.Silo().Files().Create(context.Background(), &CreateSiloFile{
			EntryID: testEntryID,
			Name:    testFileName,
		})
		require.Error(t, err)
		assert.EqualError(t, err, "missing data, or sha256, mime and size")
	})

	t.Run("registers the file when hash, mime and size are given", func(t *testing.T) {
		var body map[string]any
		responder := func(req *http.Request) (*http.Response, error) {
			raw, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &body))
			return fileResponse(`{"id":"file-id","stored":false}`), nil
		}
		c := New()
		c.conn = resty.NewWithClient(testy.HTTPClient(responder))

		f, err := c.Silo().Files().Create(context.Background(), &CreateSiloFile{
			ID:      testFileID,
			EntryID: testEntryID,
			Name:    testFileName,
			SHA256:  "27a0b656df99dc32124b4c49f2f3c35025f0a7453f289cbaa0701435cfcf4e28",
			MIME:    testFileMIME,
			Size:    12,
		})
		require.NoError(t, err)
		assert.Equal(t, testFileID, f.ID)
		assert.False(t, f.Stored)

		// The point of this path is that the bytes do not travel inline.
		_, present := body["data"]
		assert.False(t, present, "data must be omitted when not provided")
	})
}

func TestSiloFilesCreateAndUpload(t *testing.T) {
	data := []byte("<Invoice>hello</Invoice>")
	sum := sha256.Sum256(data)
	wantHash := hex.EncodeToString(sum[:])

	var calls []string
	var meta map[string]any
	var uploaded []byte
	var uploadMIME string

	responder := func(req *http.Request) (*http.Response, error) {
		calls = append(calls, req.Method+" "+req.URL.Path)
		raw, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		switch req.URL.Path {
		case testFilePath:
			require.NoError(t, json.Unmarshal(raw, &meta))
			return fileResponse(`{"id":"file-id","stored":false}`), nil
		case testDataPath:
			uploaded = raw
			uploadMIME = req.Header.Get("Content-Type")
			return fileResponse(`{"id":"file-id","stored":true}`), nil
		}
		t.Fatalf("unexpected path %s", req.URL.Path)
		return nil, nil
	}

	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(responder))

	f, err := c.Silo().Files().CreateAndUpload(context.Background(), &CreateSiloFile{
		ID:       testFileID,
		EntryID:  testEntryID,
		Name:     testFileName,
		Key:      "ubl",
		Category: FileCategoryFormat,
		MIME:     "application/xml",
	}, data)
	require.NoError(t, err)
	assert.True(t, f.Stored)

	assert.Equal(t, []string{
		"PUT " + testFilePath,
		"PUT " + testDataPath,
	}, calls)

	// The registration call carries only the details, never the payload.
	assert.Equal(t, wantHash, meta["sha256"])
	assert.Equal(t, float64(len(data)), meta["size"])
	assert.Equal(t, testFileMIME, meta["mime"])
	assert.Equal(t, "ubl", meta["key"])
	_, present := meta["data"]
	assert.False(t, present, "data must not be sent inline")

	// The data call carries the raw bytes, not JSON or base64.
	assert.Equal(t, data, uploaded)
	assert.Equal(t, testFileMIME, uploadMIME)
}

func TestSiloFilesCreateAlwaysStreamsData(t *testing.T) {
	// Size makes no difference: the payload never travels inline, so neither
	// the API nor the silo has to hold a whole file in memory.
	for _, tt := range []struct {
		name string
		data []byte
	}{
		{"a small payload", bytes.Repeat([]byte("x"), 16)},
		{"a payload past the old inline limit", bytes.Repeat([]byte("x"), 2*1024*1024)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var meta map[string]any
			var uploaded int
			responder := func(req *http.Request) (*http.Response, error) {
				calls = append(calls, req.URL.Path)
				raw, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				if strings.HasSuffix(req.URL.Path, "/data") {
					uploaded = len(raw)
					return fileResponse(`{"id":"file-id","stored":true}`), nil
				}
				require.NoError(t, json.Unmarshal(raw, &meta))
				return fileResponse(`{"id":"file-id","stored":false}`), nil
			}
			c := New()
			c.conn = resty.NewWithClient(testy.HTTPClient(responder))

			f, err := c.Silo().Files().Create(context.Background(), &CreateSiloFile{
				ID:      testFileID,
				EntryID: testEntryID,
				Name:    testFileName,
				MIME:    testFileMIME,
				Data:    tt.data,
			})
			require.NoError(t, err)
			assert.True(t, f.Stored)

			assert.Equal(t, []string{testFilePath, testDataPath}, calls)

			_, present := meta["data"]
			assert.False(t, present, "data must never travel inline")
			assert.Equal(t, float64(len(tt.data)), meta["size"])
			assert.Equal(t, len(tt.data), uploaded)
		})
	}
}

func TestSiloFilesCreateAndUploadSkipsStoredContent(t *testing.T) {
	// The silo hands back the existing file when it already holds content with
	// the same hash, so sending the payload again would be wasted work.
	var calls []string
	responder := func(req *http.Request) (*http.Response, error) {
		calls = append(calls, req.URL.Path)
		return fileResponse(`{"id":"file-id","stored":true}`), nil
	}
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(responder))

	f, err := c.Silo().Files().CreateAndUpload(context.Background(), &CreateSiloFile{
		ID:      testFileID,
		EntryID: testEntryID,
		Name:    testFileName,
	}, []byte("already stored"))
	require.NoError(t, err)
	assert.True(t, f.Stored)
	assert.Equal(t, []string{testFilePath}, calls)
}

func TestSiloFilesUploadData(t *testing.T) {
	t.Run("defaults the content type when none is given", func(t *testing.T) {
		var mime string
		responder := func(req *http.Request) (*http.Response, error) {
			mime = req.Header.Get("Content-Type")
			return fileResponse(`{"id":"file-id","stored":true}`), nil
		}
		c := New()
		c.conn = resty.NewWithClient(testy.HTTPClient(responder))

		_, err := c.Silo().Files().UploadData(context.Background(), &UploadSiloFileData{
			ID:      testFileID,
			EntryID: testEntryID,
			Data:    bytes.NewReader([]byte("data")),
		})
		require.NoError(t, err)
		assert.Equal(t, "application/octet-stream", mime)
	})

	t.Run("rejects a missing reader", func(t *testing.T) {
		c := New()
		_, err := c.Silo().Files().UploadData(context.Background(), &UploadSiloFileData{
			ID:      testFileID,
			EntryID: testEntryID,
		})
		require.Error(t, err)
		assert.EqualError(t, err, "missing data")
	})
}
