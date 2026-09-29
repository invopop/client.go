package invopop

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceList(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"list":[{"id":"series-1","name":"Sales"}]}`)
	c := testClient(t, rec)

	col, err := c.Sequence().List(context.Background())
	require.NoError(t, err)
	require.Len(t, col.List, 1)
	assert.Equal(t, "series-1", col.List[0].ID)
	assert.Equal(t, "Sales", col.List[0].Name)

	req := rec.last()
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/sequence/v1/series", req.URL.Path)
}

func TestSequenceFetch(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"id":"series-1","name":"Sales","prefix":"INV-","padding":5,"last_index":42}`)
	c := testClient(t, rec)

	s, err := c.Sequence().Fetch(context.Background(), "series-1")
	require.NoError(t, err)
	assert.Equal(t, "series-1", s.ID)
	assert.Equal(t, "INV-", s.Prefix)
	assert.Equal(t, int32(5), s.Padding)
	assert.Equal(t, int64(42), s.LastIndex)
	assert.Equal(t, "/sequence/v1/series/series-1", rec.last().URL.Path)
}

func TestSequenceCreate(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"id":"series-1"}`)
	c := testClient(t, rec)

	_, err := c.Sequence().Create(context.Background(), &CreateSeries{
		ID:     "series-1",
		Name:   "Sales",
		Prefix: "INV-",
		Start:  1,
	})
	require.NoError(t, err)

	req := rec.last()
	assert.Equal(t, http.MethodPut, req.Method)
	assert.Equal(t, "/sequence/v1/series/series-1", req.URL.Path)

	body := rec.lastBody()
	assert.Contains(t, body, `"name":"Sales"`)
	assert.Contains(t, body, `"prefix":"INV-"`)
}

func TestSequenceEntries(t *testing.T) {
	ctx := context.Background()

	t.Run("fetch entry", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1","code":"INV-00042"}`)
		c := testClient(t, rec)

		e, err := c.Sequence().FetchEntry(ctx, "series-1", testEntryID)
		require.NoError(t, err)
		assert.Equal(t, "INV-00042", e.Code)
		assert.Equal(t, "/sequence/v1/series/series-1/entries/entry-1", rec.last().URL.Path)
	})

	t.Run("create entry", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1","code":"INV-00042"}`)
		c := testClient(t, rec)

		e, err := c.Sequence().CreateEntry(ctx, "series-1", &CreateSeriesEntry{
			ID:   testEntryID,
			Meta: map[string]string{"name": "John Doe"},
		})
		require.NoError(t, err)
		assert.Equal(t, "INV-00042", e.Code)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/sequence/v1/series/series-1/entries/entry-1", req.URL.Path)
		assert.Contains(t, rec.lastBody(), `"name":"John Doe"`)
	})

	t.Run("with a conflict", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusConflict, `{"code":"duplicate","message":"already exists"}`))

		_, err := c.Sequence().CreateEntry(ctx, "series-1", &CreateSeriesEntry{ID: testEntryID})
		require.Error(t, err)
		assert.True(t, IsConflict(err))
	})
}
