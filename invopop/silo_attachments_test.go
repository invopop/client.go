package invopop

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/invopop/gobl/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			req:  &CreateSiloFile{Name: testFileName, Data: []byte("x")},
			err:  "missing entry_id",
		},
		{
			name: "without data",
			req:  &CreateSiloFile{EntryID: testEntryID, Name: testFileName},
			err:  errMissingData,
		},
		{
			name: "with empty data",
			req:  &CreateSiloFile{EntryID: testEntryID, Name: testFileName, Data: []byte{}},
			err:  errMissingData,
		},
		{
			name: "without a name",
			req:  &CreateSiloFile{EntryID: testEntryID, Data: []byte("x")},
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

func TestSiloFilesCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("with an explicit id", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"file-1","name":"invoice.pdf","stored":true}`)
		c := testClient(t, rec)

		f, err := c.Silo().Files().Create(ctx, &CreateSiloFile{
			ID:       "file-1",
			EntryID:  testEntryID,
			Name:     "invoice.pdf",
			Category: FileCategoryFormat,
			Data:     []byte("data"),
		})
		require.NoError(t, err)
		assert.Equal(t, "file-1", f.ID)
		assert.True(t, f.Stored)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/files/file-1", req.URL.Path)

		body := rec.lastBody()
		assert.Contains(t, body, `"name":"invoice.pdf"`)
		assert.Contains(t, body, `"category":"format"`)
		assert.Contains(t, body, `"data":"ZGF0YQ=="`, "data should be base64 encoded")
	})

	t.Run("assigns a uuid when none is provided", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"file-1"}`)
		c := testClient(t, rec)

		req := &CreateSiloFile{
			EntryID: testEntryID,
			Name:    "invoice.pdf",
			Data:    []byte("data"),
		}
		_, err := c.Silo().Files().Create(ctx, req)
		require.NoError(t, err)

		require.NotEmpty(t, req.ID, "the request should be updated with the new ID")
		parsed, err := uuid.Parse(req.ID)
		require.NoError(t, err, "should assign a valid UUID")
		assert.Equal(t, 7, int(parsed.Version()))
		assert.Equal(t, "/silo/v1/entries/entry-1/files/"+req.ID, rec.last().URL.Path)
	})
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
