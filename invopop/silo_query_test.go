package invopop

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiloEntriesQuery(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		req   *QuerySiloEntries
		query string
	}{
		{name: "with only a schema", req: &QuerySiloEntries{Schema: testSchema}, query: "schema=bill%2Finvoice"},
		{
			name:  "with a filter containing an ampersand",
			req:   &QuerySiloEntries{Schema: testSchema, Filter: `customer_name = "A & B"`},
			query: "filter=customer_name+%3D+%22A+%26+B%22&schema=bill%2Finvoice",
		},
		{
			name: "with every option",
			req: &QuerySiloEntries{
				Schema:  testSchema,
				Filter:  `state = "sent"`,
				OrderBy: "issue_date desc",
				Limit:   20,
				Cursor:  testCursor,
				Fields:  "id,code",
				Count:   true,
			},
			query: "count=true&cursor=abc&fields=id%2Ccode&filter=state+%3D+%22sent%22&limit=20&order_by=issue_date+desc&schema=bill%2Finvoice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(http.StatusOK, `{"list":[]}`)
			c := testClient(t, rec)

			_, err := c.Silo().Entries().Query(ctx, tt.req)
			require.NoError(t, err)

			req := rec.last()
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/silo/v1/entries/query", req.URL.Path)
			assert.Equal(t, tt.query, req.URL.RawQuery)
		})
	}

	t.Run("parses the response", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{
			"list":[{"id":"entry-1","folder":"sales"}],
			"schema":"bill/invoice",
			"limit":50,
			"next_cursor":"def",
			"count":42,
			"count_capped":true,
			"watermark":"2026-01-01T00:00:00.000Z",
			"notes":["no index covers this filter"]
		}`)
		c := testClient(t, rec)

		col, err := c.Silo().Entries().Query(ctx, &QuerySiloEntries{Schema: testSchema})
		require.NoError(t, err)
		require.Len(t, col.List, 1)
		assert.Equal(t, testEntryID, col.List[0].ID)
		assert.Equal(t, "def", col.NextCursor)
		assert.Equal(t, int32(42), col.Count)
		assert.True(t, col.CountCapped)
		assert.Equal(t, "2026-01-01T00:00:00.000Z", col.Watermark)
		assert.Equal(t, []string{"no index covers this filter"}, col.Notes)
	})
}

func TestSiloEntriesCount(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{"schema":"bill/invoice","count":7}`)
	c := testClient(t, rec)

	out, err := c.Silo().Entries().Count(ctx, &CountSiloEntries{Schema: testSchema, Filter: `state = "sent"`})
	require.NoError(t, err)
	assert.Equal(t, int32(7), out.Count)

	req := rec.last()
	assert.Equal(t, "/silo/v1/entries/query/count", req.URL.Path)
	assert.Equal(t, "filter=state+%3D+%22sent%22&schema=bill%2Finvoice", req.URL.RawQuery)
}
