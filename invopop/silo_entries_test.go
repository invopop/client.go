package invopop

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/invopop/client.go/pkg/snippets"

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

func TestSiloEntriesList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		req   *FindSiloEntries
		query string
	}{
		{name: "with no request", req: nil, query: ""},
		{name: "with an empty request", req: new(FindSiloEntries), query: ""},
		{name: "with a folder", req: &FindSiloEntries{Folder: testFolder}, query: "folder=sales"},
		{name: "with a limit", req: &FindSiloEntries{Limit: 20}, query: "limit=20"},
		{name: "with a cursor", req: &FindSiloEntries{Cursor: testCursor}, query: "cursor=abc"},
		{
			name:  "with every filter",
			req:   &FindSiloEntries{Folder: testFolder, Limit: 20, Cursor: testCursor, CreatedAt: testCreatedAt},
			query: "created_at=2023-08-02T00%3A00%3A00.000Z&cursor=abc&folder=sales&limit=20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(http.StatusOK, `{"list":[]}`)
			c := testClient(t, rec)

			col, err := c.Silo().Entries().List(ctx, tt.req)
			require.NoError(t, err)
			require.NotNil(t, col)

			req := rec.last()
			assert.Equal(t, "/silo/v1/entries", req.URL.Path)
			assert.Equal(t, tt.query, req.URL.RawQuery)
		})
	}
}

func TestSiloEntriesFetch(t *testing.T) {
	ctx := context.Background()

	t.Run("by id", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1","folder":"sales"}`)
		c := testClient(t, rec)

		e, err := c.Silo().Entries().Fetch(ctx, testEntryID)
		require.NoError(t, err)
		assert.Equal(t, testEntryID, e.ID)
		assert.Equal(t, testFolder, e.Folder)
		assert.Equal(t, "/silo/v1/entries/entry-1", rec.last().URL.Path)
	})

	t.Run("by key", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1","key":"invoice-101"}`)
		c := testClient(t, rec)

		e, err := c.Silo().Entries().FetchByKey(ctx, "invoice-101")
		require.NoError(t, err)
		assert.Equal(t, "invoice-101", e.Key)
		assert.Equal(t, "/silo/v1/entries/key/invoice-101", rec.last().URL.Path)
	})
}

func TestSiloEntriesCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("without an id posts", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1"}`)
		c := testClient(t, rec)

		_, err := c.Silo().Entries().Create(ctx, &CreateSiloEntry{Folder: testFolder})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/silo/v1/entries", req.URL.Path)
	})

	t.Run("with an id puts", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1"}`)
		c := testClient(t, rec)

		_, err := c.Silo().Entries().Create(ctx, &CreateSiloEntry{ID: testEntryID})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1", req.URL.Path)
	})
}

func TestSiloEntriesUpdate(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"id":"entry-1"}`)
	c := testClient(t, rec)

	_, err := c.Silo().Entries().Update(context.Background(), &UpdateSiloEntry{
		ID:          testEntryID,
		ContentType: MIMEApplicationMergePatchJSON,
		Data:        json.RawMessage(`{"doc":{"code":"NEW"}}`),
	})
	require.NoError(t, err)

	req := rec.last()
	assert.Equal(t, http.MethodPatch, req.Method)
	assert.Equal(t, "/silo/v1/entries/entry-1", req.URL.Path)
	assert.Contains(t, rec.lastBody(), `"content_type":"application/merge-patch+json"`)
}

const testEnvelope = `{
	"$schema": "https://gobl.org/draft-0/envelope",
	"head": {
		"uuid": "0192cd23-5c2a-7000-8000-000000000001",
		"dig": {"alg":"sha256","val":"a1b2c3"}
	},
	"doc": {
		"$schema": "https://gobl.org/draft-0/bill/invoice",
		"code": "SAMPLE-001",
		"currency": "EUR",
		"issue_date": "2024-11-01",
		"supplier": {"name": "Test Supplier"}
	}
}`

func TestSiloEntryEnvelope(t *testing.T) {
	t.Run("parses the envelope data", func(t *testing.T) {
		se := &SiloEntry{Data: json.RawMessage(testEnvelope)}
		env, err := se.Envelope()
		require.NoError(t, err)
		require.NotNil(t, env)

		require.NotNil(t, env.Head)
		require.NotNil(t, env.Head.Digest)
		assert.Equal(t, "a1b2c3", env.Head.Digest.Value)
		require.NotNil(t, env.Document)
		assert.Equal(t, "https://gobl.org/draft-0/bill/invoice", env.Document.Schema.String())
	})

	t.Run("with invalid data", func(t *testing.T) {
		se := &SiloEntry{Data: json.RawMessage(`not json`)}
		env, err := se.Envelope()
		assert.Error(t, err)
		assert.Nil(t, env)
	})

	t.Run("with no data", func(t *testing.T) {
		se := new(SiloEntry)
		_, err := se.Envelope()
		assert.Error(t, err, "an empty payload is not a valid envelope")
	})
}

func TestSiloEntryDataEnvelope(t *testing.T) {
	t.Run("parses the versioned envelope data", func(t *testing.T) {
		sd := &SiloEntryData{Version: "v1", Data: json.RawMessage(testEnvelope)}
		env, err := sd.Envelope()
		require.NoError(t, err)
		require.NotNil(t, env)
		assert.Equal(t, "https://gobl.org/draft-0/bill/invoice", env.Document.Schema.String())
	})

	t.Run("with invalid data", func(t *testing.T) {
		sd := &SiloEntryData{Data: json.RawMessage(`not json`)}
		_, err := sd.Envelope()
		assert.Error(t, err)
	})
}

func TestSiloEntrySnippet(t *testing.T) {
	t.Run("with an invoice", func(t *testing.T) {
		se := &SiloEntry{
			DocSchema:   "https://gobl.org/draft-0/bill/invoice",
			SnippetData: json.RawMessage(`{"code":"0123","currency":"EUR"}`),
		}
		s := se.Snippet()
		require.NotNil(t, s)

		inv, ok := s.(*snippets.BillInvoice)
		require.True(t, ok)
		assert.Equal(t, "0123", inv.Code)
	})

	t.Run("with an unknown schema", func(t *testing.T) {
		se := &SiloEntry{DocSchema: "https://gobl.org/draft-0/pay/advance"}
		assert.Nil(t, se.Snippet())
	})

	t.Run("with no snippet data", func(t *testing.T) {
		se := new(SiloEntry)
		assert.Nil(t, se.Snippet())
	})
}

func TestSiloGOBL(t *testing.T) {
	ctx := context.Background()

	t.Run("build", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"data":{"$schema":"https://gobl.org/draft-0/envelope"}}`)
		c := testClient(t, rec)

		out, err := c.Silo().GOBL().Build(ctx, &BuildGOBL{
			Data:    json.RawMessage(`{"code":"0123"}`),
			Envelop: true,
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"$schema":"https://gobl.org/draft-0/envelope"}`, string(out.Data))

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/silo/v1/gobl/build", req.URL.Path)
		assert.Contains(t, rec.lastBody(), `"envelop":true`)
	})

	t.Run("sign", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"data":{}}`)
		c := testClient(t, rec)

		_, err := c.Silo().GOBL().Sign(ctx, &SignGOBL{Data: json.RawMessage(`{}`)})
		require.NoError(t, err)
		assert.Equal(t, "/silo/v1/gobl/sign", rec.last().URL.Path)
	})

	t.Run("with a validation error", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusUnprocessableEntity, `{
			"message": "validation failed",
			"fields": {"supplier": {"tax_id": "cannot be blank"}}
		}`))

		_, err := c.Silo().GOBL().Build(ctx, &BuildGOBL{Data: json.RawMessage(`{}`)})
		require.Error(t, err)

		re := AsResponseError(err)
		require.NotNil(t, re)
		assert.Equal(t, "cannot be blank", re.Fields.Get("supplier.tax_id").Message())
	})
}
