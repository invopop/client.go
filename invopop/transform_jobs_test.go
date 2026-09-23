package invopop

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobDone(t *testing.T) {
	t.Run("when not completed", func(t *testing.T) {
		j := new(Job)
		done, err := j.Done()
		assert.False(t, done)
		assert.NoError(t, err)
	})

	t.Run("when not completed with intents present", func(t *testing.T) {
		j := &Job{
			Intents: []*JobIntent{
				{StepID: testStepID, Events: []*JobIntentEvent{{Status: "OK"}}},
			},
		}
		done, err := j.Done()
		assert.False(t, done, "completed_at should determine completion")
		assert.NoError(t, err)
	})

	t.Run("when completed successfully", func(t *testing.T) {
		j := &Job{
			CompletedAt: testCompletedAt,
			Intents: []*JobIntent{
				{
					StepID: testStepID,
					Events: []*JobIntentEvent{
						{Index: 0, Status: "RUN"},
						{Index: 1, Status: "OK"},
					},
				},
			},
		}
		done, err := j.Done()
		assert.True(t, done)
		assert.NoError(t, err)
	})

	t.Run("when the last event failed", func(t *testing.T) {
		j := &Job{
			CompletedAt: testCompletedAt,
			Intents: []*JobIntent{
				{
					StepID: testStepID,
					Events: []*JobIntentEvent{{Status: "OK"}},
				},
				{
					StepID: "step-2",
					Events: []*JobIntentEvent{
						{Status: "RUN"},
						{Status: "KO", At: "2024-11-01T00:00:01.000Z", Message: "provider rejected"},
					},
				},
			},
		}
		done, err := j.Done()
		assert.True(t, done, "a failed job is still done")
		require.Error(t, err)
		assert.EqualError(t, err, "step step-2 failed at 2024-11-01T00:00:01.000Z: provider rejected")
	})

	t.Run("only the last intent and event are considered", func(t *testing.T) {
		// An earlier KO that was retried successfully should not surface.
		j := &Job{
			CompletedAt: testCompletedAt,
			Intents: []*JobIntent{
				{StepID: testStepID, Events: []*JobIntentEvent{{Status: "KO"}}},
				{StepID: "step-2", Events: []*JobIntentEvent{{Status: "KO"}, {Status: "OK"}}},
			},
		}
		done, err := j.Done()
		assert.True(t, done)
		assert.NoError(t, err)
	})
}

func TestJobsList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		req   *FindJobs
		query string
	}{
		{
			name:  "with no request",
			req:   nil,
			query: "",
		},
		{
			name:  "with an empty request",
			req:   new(FindJobs),
			query: "",
		},
		{
			name:  "with a limit",
			req:   &FindJobs{Limit: 20},
			query: "limit=20",
		},
		{
			name:  "with a cursor",
			req:   &FindJobs{Cursor: testCursor},
			query: "cursor=abc",
		},
		{
			name:  "with a created at filter",
			req:   &FindJobs{CreatedAt: testCreatedAt},
			query: "created_at=2023-08-02T00%3A00%3A00.000Z",
		},
		{
			name:  "with every filter",
			req:   &FindJobs{Limit: 10, Cursor: testCursor, CreatedAt: testCreatedAt},
			query: "created_at=2023-08-02T00%3A00%3A00.000Z&cursor=abc&limit=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(http.StatusOK, `{"list":[],"limit":10}`)
			c := testClient(t, rec)

			col, err := c.Transform().Jobs().List(ctx, tt.req)
			require.NoError(t, err)
			require.NotNil(t, col)

			req := rec.last()
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/transform/v1/jobs", req.URL.Path)
			assert.Equal(t, tt.query, req.URL.RawQuery)
		})
	}

	t.Run("decodes the collection", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{
			"list": [{"id":"job-1","workflow_id":"wf-1","status":"ok"}],
			"limit": 10,
			"next_cursor": "next"
		}`)
		c := testClient(t, rec)

		col, err := c.Transform().Jobs().List(ctx, nil)
		require.NoError(t, err)
		require.Len(t, col.List, 1)
		assert.Equal(t, "job-1", col.List[0].ID)
		assert.Equal(t, "ok", col.List[0].Status)
		assert.Equal(t, "next", col.NextCursor)
	})
}

func TestJobsCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("without an id posts", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"job-1"}`)
		c := testClient(t, rec)

		j, err := c.Transform().Jobs().Create(ctx, &CreateJob{WorkflowID: "wf-1"})
		require.NoError(t, err)
		assert.Equal(t, "job-1", j.ID)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/transform/v1/jobs", req.URL.Path)
		assert.Contains(t, rec.lastBody(), `"workflow_id":"wf-1"`)
	})

	t.Run("with an id puts", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"job-1"}`)
		c := testClient(t, rec)

		_, err := c.Transform().Jobs().Create(ctx, &CreateJob{ID: "job-1", WorkflowID: "wf-1"})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/transform/v1/jobs/job-1", req.URL.Path)
	})

	t.Run("with a wait adds the query parameter", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"job-1"}`)
		c := testClient(t, rec)

		_, err := c.Transform().Jobs().Create(ctx, &CreateJob{ID: "job-1", Wait: 30})
		require.NoError(t, err)
		assert.Equal(t, "wait=30", rec.last().URL.RawQuery)
	})
}

func TestJobsFetch(t *testing.T) {
	ctx := context.Background()

	t.Run("by id", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"job-1"}`)
		c := testClient(t, rec)

		j, err := c.Transform().Jobs().Fetch(ctx, "job-1")
		require.NoError(t, err)
		assert.Equal(t, "job-1", j.ID)
		assert.Equal(t, "/transform/v1/jobs/job-1", rec.last().URL.Path)
	})

	t.Run("by key", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"job-1","key":"my-key"}`)
		c := testClient(t, rec)

		j, err := c.Transform().Jobs().FetchByKey(ctx, "my-key")
		require.NoError(t, err)
		assert.Equal(t, "my-key", j.Key)
		assert.Equal(t, "/transform/v1/jobs/key/my-key", rec.last().URL.Path)
	})

	t.Run("with a not found response", func(t *testing.T) {
		c := testClient(t, newRecorder(http.StatusNotFound, `{"message":"unknown job"}`))
		_, err := c.Transform().Jobs().Fetch(ctx, "missing")
		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})
}

func TestJobsUpdateIntent(t *testing.T) {
	ctx := context.Background()

	t.Run("update intent", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"intent-1"}`)
		c := testClient(t, rec)

		in, err := c.Transform().Jobs().UpdateIntent(ctx, &UpdateIntent{
			ID:     "intent-1",
			Status: "POKE",
		})
		require.NoError(t, err)
		assert.Equal(t, "intent-1", in.ID)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/transform/v1/jobs/intents", req.URL.Path)
	})

	t.Run("poke by ref", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"intent-1"}`)
		c := testClient(t, rec)

		_, err := c.Transform().Jobs().PokeByRef(ctx, "my-ref")
		require.NoError(t, err)

		body := rec.lastBody()
		assert.Contains(t, body, `"ref":"my-ref"`)
		assert.Contains(t, body, `"status":"POKE"`)
	})
}
