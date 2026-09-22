package invopop

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiloMetaValidation(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{}`)
	svc := testClient(t, rec).Silo().Meta()

	tests := []struct {
		name string
		call func() error
		err  string
	}{
		{
			name: "fetch without an entry id",
			call: func() error { _, err := svc.Fetch(ctx, "", "key"); return err },
			err:  "missing entry ID",
		},
		{
			name: "fetch without a key",
			call: func() error { _, err := svc.Fetch(ctx, testEntryID, ""); return err },
			err:  errMissingKey,
		},
		{
			name: "fetch by ref without a key",
			call: func() error { _, err := svc.FetchByRef(ctx, "", "ref"); return err },
			err:  errMissingKey,
		},
		{
			name: "fetch by ref without a ref",
			call: func() error { _, err := svc.FetchByRef(ctx, "key", ""); return err },
			err:  "missing ref",
		},
		{
			name: "fetch by owner and ref without a key",
			call: func() error { _, err := svc.FetchByOwnerAndRef(ctx, "", "ref"); return err },
			err:  errMissingKey,
		},
		{
			name: "fetch by owner and ref without a ref",
			call: func() error { _, err := svc.FetchByOwnerAndRef(ctx, "key", ""); return err },
			err:  "missing ref",
		},
		{
			name: "upsert without a key",
			call: func() error { _, err := svc.Upsert(ctx, &UpsertSiloMeta{EntryID: testEntryID}); return err },
			err:  errMissingKey,
		},
		{
			name: "delete without an entry id",
			call: func() error { _, err := svc.Delete(ctx, "", "key"); return err },
			err:  "missing entry ID",
		},
		{
			name: "delete without a key",
			call: func() error { _, err := svc.Delete(ctx, testEntryID, ""); return err },
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

func TestSiloMetaRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("fetch", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1:src:key","key":"key","value":{"a":1}}`)
		c := testClient(t, rec)

		m, err := c.Silo().Meta().Fetch(ctx, testEntryID, "key")
		require.NoError(t, err)
		assert.Equal(t, "key", m.Key)
		assert.JSONEq(t, `{"a":1}`, string(m.Value))

		req := rec.last()
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/meta/key", req.URL.Path)
	})

	t.Run("fetch by ref", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"ref":"my-ref"}`)
		c := testClient(t, rec)

		m, err := c.Silo().Meta().FetchByRef(ctx, "key", "my-ref")
		require.NoError(t, err)
		assert.Equal(t, "my-ref", m.Ref)
		assert.Equal(t, "/silo/v1/entries/meta/key/my-ref", rec.last().URL.Path)
		assert.Empty(t, rec.last().URL.RawQuery)
	})

	t.Run("fetch by owner and ref", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"ref":"my-ref"}`)
		c := testClient(t, rec)

		_, err := c.Silo().Meta().FetchByOwnerAndRef(ctx, "key", "my-ref")
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, "/silo/v1/entries/meta/key/my-ref", req.URL.Path)
		assert.Equal(t, "owned=true", req.URL.RawQuery, "should filter by the token's owner")
	})

	t.Run("upsert", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1:src:key"}`)
		c := testClient(t, rec)

		_, err := c.Silo().Meta().Upsert(ctx, &UpsertSiloMeta{
			EntryID: testEntryID,
			Key:     "key",
			Ref:     "my-ref",
			Value:   json.RawMessage(`{"a":1}`),
			Indexed: true,
		})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/meta/key", req.URL.Path)

		body := rec.lastBody()
		assert.Contains(t, body, `"ref":"my-ref"`)
		assert.Contains(t, body, `"indexed":true`)
		assert.NotContains(t, body, `"entry_id"`, "entry id travels in the path, not the body")
	})

	t.Run("upsert without an entry id", func(t *testing.T) {
		// Permitted: the ref alone may identify the row.
		rec := newRecorder(http.StatusOK, `{}`)
		c := testClient(t, rec)

		_, err := c.Silo().Meta().Upsert(ctx, &UpsertSiloMeta{Key: "key", Ref: "my-ref"})
		require.NoError(t, err)
		assert.Equal(t, "/silo/v1/entries/meta/key", rec.last().URL.Path)
	})

	t.Run("delete", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{}`)
		c := testClient(t, rec)

		_, err := c.Silo().Meta().Delete(ctx, testEntryID, "key")
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodDelete, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/meta/key", req.URL.Path)
	})
}
