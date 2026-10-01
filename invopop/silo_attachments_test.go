package invopop

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

const (
	errMissingFileData = "missing data, or sha256, mime and size"

	testFileID   = "file-id"
	testFileMIME = "application/xml"
	testFilePath = "/silo/v1/entries/" + testEntryID + "/files/" + testFileID
	testDataPath = testFilePath + "/data"
)

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
			return jsonResponse(`{"id":"file-id","stored":false}`), nil
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

		_, present := body["data"]
		assert.False(t, present, "data must be omitted when not provided")
	})
}

func TestSiloFilesCreateUploadsDataSeparately(t *testing.T) {
	data := []byte("<Invoice>hello</Invoice>")
	wantHash := dsig.NewSHA256Digest(data).Value

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
			return jsonResponse(`{"id":"file-id","stored":false}`), nil
		case testDataPath:
			uploaded = raw
			uploadMIME = req.Header.Get("Content-Type")
			return jsonResponse(`{"id":"file-id","stored":true}`), nil
		}
		t.Fatalf("unexpected path %s", req.URL.Path)
		return nil, nil
	}

	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(responder))

	f, err := c.Silo().Files().Create(context.Background(), &CreateSiloFile{
		ID:       testFileID,
		EntryID:  testEntryID,
		Name:     testFileName,
		Key:      "ubl",
		Category: FileCategoryFormat,
		MIME:     testFileMIME,
		Data:     bytes.NewReader(data),
	})
	require.NoError(t, err)
	assert.True(t, f.Stored)

	assert.Equal(t, []string{
		"PUT " + testFilePath,
		"PUT " + testDataPath,
	}, calls)

	// registration carries details only
	assert.Equal(t, wantHash, meta["sha256"])
	assert.Equal(t, float64(len(data)), meta["size"])
	assert.Equal(t, testFileMIME, meta["mime"])
	assert.Equal(t, "ubl", meta["key"])
	_, present := meta["data"]
	assert.False(t, present, "data must not be sent inline")

	// data call carries raw bytes
	assert.Equal(t, data, uploaded)
	assert.Equal(t, testFileMIME, uploadMIME)
}

func TestSiloFilesCreateRetriesFailedUpload(t *testing.T) {
	data := []byte("<Invoice>hello</Invoice>")

	tests := []struct {
		name    string
		content io.Reader
	}{
		{name: "with a seekable reader", content: bytes.NewReader(data)},
		// Held in memory on the first attempt, and kept on the request.
		{name: "with a reader that cannot seek", content: readerOnly{bytes.NewReader(data)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var uploads [][]byte
			failUpload := true

			responder := func(req *http.Request) (*http.Response, error) {
				calls = append(calls, req.Method+" "+req.URL.Path)
				raw, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				switch req.URL.Path {
				case testFilePath:
					return jsonResponse(`{"id":"file-id","stored":false}`), nil
				case testDataPath:
					uploads = append(uploads, raw)
					if failUpload {
						failUpload = false
						res := jsonResponse(`{"message":"unavailable"}`)
						res.StatusCode = http.StatusServiceUnavailable
						return res, nil
					}
					return jsonResponse(`{"id":"file-id","stored":true}`), nil
				}
				t.Fatalf("unexpected path %s", req.URL.Path)
				return nil, nil
			}

			c := New()
			c.conn = resty.NewWithClient(testy.HTTPClient(responder))

			req := &CreateSiloFile{
				ID:      testFileID,
				EntryID: testEntryID,
				Name:    testFileName,
				MIME:    testFileMIME,
				Data:    tt.content,
			}
			_, err := c.Silo().Files().Create(context.Background(), req)
			require.Error(t, err)

			f, err := c.Silo().Files().Create(context.Background(), req)
			require.NoError(t, err)
			assert.True(t, f.Stored)

			assert.Equal(t, []string{
				"PUT " + testFilePath,
				"PUT " + testDataPath,
				"PUT " + testFilePath,
				"PUT " + testDataPath,
			}, calls)
			assert.Equal(t, [][]byte{data, data}, uploads, "the retry must send the contents again")
		})
	}
}

// readerOnly hides any Seek the underlying reader has.
type readerOnly struct{ r io.Reader }

func (r readerOnly) Read(p []byte) (int, error) { return r.r.Read(p) }

func TestSiloFilesCreateFromReader(t *testing.T) {
	data := []byte("<Invoice>streamed</Invoice>")
	wantHash := dsig.NewSHA256Digest(data).Value

	run := func(t *testing.T, content io.Reader, req *CreateSiloFile) (map[string]any, []byte, []string) {
		t.Helper()
		var calls []string
		var meta map[string]any
		var uploaded []byte

		responder := func(r *http.Request) (*http.Response, error) {
			calls = append(calls, r.Method+" "+r.URL.Path)
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			switch r.URL.Path {
			case testFilePath:
				require.NoError(t, json.Unmarshal(raw, &meta))
				return jsonResponse(`{"id":"file-id","stored":false}`), nil
			case testDataPath:
				uploaded = raw
				return jsonResponse(`{"id":"file-id","stored":true}`), nil
			}
			t.Fatalf("unexpected path %s", r.URL.Path)
			return nil, nil
		}

		c := New()
		c.conn = resty.NewWithClient(testy.HTTPClient(responder))
		req.ID, req.EntryID, req.Name, req.Data = testFileID, testEntryID, testFileName, content

		f, err := c.Silo().Files().Create(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, f.Stored)
		return meta, uploaded, calls
	}

	t.Run("measures and rewinds a seekable reader", func(t *testing.T) {
		meta, uploaded, calls := run(t, bytes.NewReader(data), &CreateSiloFile{})
		assert.Equal(t, []string{"PUT " + testFilePath, "PUT " + testDataPath}, calls)
		assert.Equal(t, wantHash, meta["sha256"])
		assert.Equal(t, float64(len(data)), meta["size"])
		assert.NotEmpty(t, meta["mime"], "mime should be detected from the contents")
		assert.Equal(t, data, uploaded, "the reader must be wound back before sending")
	})

	t.Run("holds a reader that cannot seek", func(t *testing.T) {
		meta, uploaded, _ := run(t, readerOnly{bytes.NewReader(data)}, &CreateSiloFile{})
		assert.Equal(t, wantHash, meta["sha256"])
		assert.Equal(t, float64(len(data)), meta["size"])
		assert.Equal(t, data, uploaded)
	})

	t.Run("sends a described reader straight through", func(t *testing.T) {
		// Once-through and undescribed, this would arrive empty: whatever read
		// it to find the hash would have drained it.
		meta, uploaded, _ := run(t, readerOnly{bytes.NewReader(data)}, &CreateSiloFile{
			SHA256: wantHash,
			Size:   int32(len(data)),
			MIME:   testFileMIME,
		})
		assert.Equal(t, testFileMIME, meta["mime"])
		assert.Equal(t, data, uploaded)
	})
}

func TestFileSize(t *testing.T) {
	// Exercised directly: a payload over the limit is too large to build in a
	// test, and both measuring paths share this check.
	n, err := fileSize(math.MaxInt32)
	require.NoError(t, err)
	assert.Equal(t, int32(math.MaxInt32), n)

	_, err = fileSize(math.MaxInt32 + 1)
	assert.ErrorContains(t, err, "over the 2147483647 limit")
}

func TestSiloFilesCreateSkipsStoredContent(t *testing.T) {
	// The silo returns the existing file when it already holds content with the
	// same hash, so sending the payload again would be wasted.
	var calls []string
	responder := func(req *http.Request) (*http.Response, error) {
		calls = append(calls, req.URL.Path)
		return jsonResponse(`{"id":"file-id","stored":true}`), nil
	}
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(responder))

	f, err := c.Silo().Files().Create(context.Background(), &CreateSiloFile{
		ID:      testFileID,
		EntryID: testEntryID,
		Name:    testFileName,
		Data:    bytes.NewReader([]byte("already stored")),
	})
	require.NoError(t, err)
	assert.True(t, f.Stored)
	assert.Equal(t, []string{testFilePath}, calls)
}

func TestSiloFilesUploadData(t *testing.T) {
	t.Run("defaults the content type when none is given", func(t *testing.T) {
		var mime string
		responder := func(req *http.Request) (*http.Response, error) {
			mime = req.Header.Get("Content-Type")
			return jsonResponse(`{"id":"file-id","stored":true}`), nil
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

func TestSiloFilesCreateValidation(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{}`)
	svc := testClient(t, rec).Silo().Files()

	tests := []struct {
		name string
		req  *CreateSiloFile
		err  string
	}{
		{
			name: "without an entry id",
			req:  &CreateSiloFile{Name: testFileName, Data: bytes.NewReader([]byte("x"))},
			err:  "missing entry_id",
		},
		{
			name: "without data",
			req:  &CreateSiloFile{EntryID: testEntryID, Name: testFileName},
			err:  errMissingFileData,
		},
		{
			name: "with empty data",
			req:  &CreateSiloFile{EntryID: testEntryID, Name: testFileName, Data: bytes.NewReader(nil)},
			err:  errMissingFileData,
		},
		{
			name: "without a name",
			req:  &CreateSiloFile{EntryID: testEntryID, Data: bytes.NewReader([]byte("x"))},
			err:  "missing name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.req)
			assert.EqualError(t, err, tt.err)
		})
	}

	assert.Empty(t, rec.requests, "validation should fail before any request is made")
}

func TestSiloFilesCreateAssignsID(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{"id":"file-1"}`)
	c := testClient(t, rec)

	// No data, so registration is the only call.
	req := &CreateSiloFile{
		EntryID: testEntryID,
		Name:    "invoice.pdf",
		SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		MIME:    "application/pdf",
		Size:    4,
	}
	_, err := c.Silo().Files().Create(ctx, req)
	require.NoError(t, err)

	require.NotEmpty(t, req.ID, "the request should be updated with the new ID")
	parsed, err := uuid.Parse(req.ID)
	require.NoError(t, err, "should assign a valid UUID")
	assert.Equal(t, 7, int(parsed.Version()))
	assert.Equal(t, "/silo/v1/entries/"+testEntryID+"/files/"+req.ID, rec.last().URL.Path)
}

func TestSiloFilesDownload(t *testing.T) {
	ctx := context.Background()

	t.Run("without an id", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusOK, `{}`))
		_, err := c.Silo().Files().Download(ctx, testEntryID, "")
		assert.EqualError(t, err, "missing id")
	})

	t.Run("without an entry id", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusOK, `{}`))
		_, err := c.Silo().Files().Download(ctx, "", "file-1")
		assert.EqualError(t, err, "missing entry_id")
	})

	t.Run("returns the file contents", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `file contents`)
		c := testClient(t, rec)

		body, err := c.Silo().Files().Download(ctx, testEntryID, "file-1")
		require.NoError(t, err)
		defer body.Close() //nolint:errcheck

		assert.Equal(t, "/silo/v1/entries/entry-1/files/file-1", rec.last().URL.Path)

		data, err := io.ReadAll(body)
		require.NoError(t, err)
		assert.Equal(t, "file contents", string(data))
	})

	t.Run("with an error response", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusNotFound, `{"message":"gone"}`))
		_, err := c.Silo().Files().Download(ctx, testEntryID, "file-1")
		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})
}

func TestFileCategories(t *testing.T) {
	// These values are validated by the Silo service, so they must not drift.
	assert.Equal(t, "", FileCategoryDefault)
	assert.Equal(t, "format", FileCategoryFormat)
	assert.Equal(t, "request", FileCategoryRequest)
	assert.Equal(t, "response", FileCategoryResponse)
	assert.Equal(t, "agreement", FileCategoryAgreement)
	assert.Equal(t, "verification", FileCategoryVerification)
	assert.Equal(t, "attachment", FileCategoryAttachment)
}
