package invopop

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCommentID      = "entry-1:0190c2e5-0a1b-7c2d-8e3f-4a5b6c7d8e9f"
	testCommentPath    = "/silo/v1/entries/entry-1/comments/" + testCommentID
	errMissingComment  = "missing comment ID"
	testAgentID        = "agent-1"
	testCommentRespond = `{"id":"` + testCommentID + `","entry_id":"entry-1","body":"Check the total","author":{"kind":"user","id":"u1"}}`
)

func TestSiloCommentsValidation(t *testing.T) {
	ctx := context.Background()
	rec := newRecorder(http.StatusOK, `{}`)
	svc := testClient(t, rec).Silo().Comments()

	tests := []struct {
		name string
		call func() error
		err  string
	}{
		{
			name: "list without a request",
			call: func() error { _, err := svc.List(ctx, nil); return err },
			err:  errMissingEntryID,
		},
		{
			name: "list without an entry id",
			call: func() error { _, err := svc.List(ctx, &FindSiloComments{}); return err },
			err:  errMissingEntryID,
		},
		{
			name: "fetch without an entry id",
			call: func() error { _, err := svc.Fetch(ctx, "", testCommentID); return err },
			err:  errMissingEntryID,
		},
		{
			name: "fetch without an id",
			call: func() error { _, err := svc.Fetch(ctx, testEntryID, ""); return err },
			err:  errMissingComment,
		},
		{
			name: "create without an entry id",
			call: func() error { _, err := svc.Create(ctx, &CreateSiloComment{Body: "x"}); return err },
			err:  errMissingEntryID,
		},
		{
			name: "create without a body",
			call: func() error { _, err := svc.Create(ctx, &CreateSiloComment{EntryID: testEntryID}); return err },
			err:  errMissingBody,
		},
		{
			name: "update without an id",
			call: func() error {
				_, err := svc.Update(ctx, &UpdateSiloComment{EntryID: testEntryID, Body: "x"})
				return err
			},
			err: errMissingComment,
		},
		{
			name: "update without a body",
			call: func() error {
				_, err := svc.Update(ctx, &UpdateSiloComment{EntryID: testEntryID, ID: testCommentID})
				return err
			},
			err: errMissingBody,
		},
		{
			name: "resolve without an id",
			call: func() error { _, err := svc.Resolve(ctx, &ActOnSiloComment{EntryID: testEntryID}); return err },
			err:  errMissingComment,
		},
		{
			name: "reopen without an entry id",
			call: func() error { _, err := svc.Reopen(ctx, &ActOnSiloComment{ID: testCommentID}); return err },
			err:  errMissingEntryID,
		},
		{
			name: "delete without a request",
			call: func() error { _, err := svc.Delete(ctx, nil); return err },
			err:  errMissingEntryID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.call(), tt.err)
		})
	}

	assert.Empty(t, rec.requests, "validation should fail before any request is made")
}

func TestSiloCommentsRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("list", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"list":[{"id":"c1","body":"a","replies":[{"id":"c2","reply_to":"c1","body":"b"}]}]}`)
		c := testClient(t, rec)

		col, err := c.Silo().Comments().List(ctx, &FindSiloComments{EntryID: testEntryID})
		require.NoError(t, err)
		require.Len(t, col.List, 1)
		require.Len(t, col.List[0].Replies, 1)
		assert.Equal(t, "c1", col.List[0].Replies[0].ReplyTo)

		req := rec.last()
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/comments", req.URL.Path)
		assert.Empty(t, req.URL.RawQuery)
	})

	t.Run("list with filters", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"list":[]}`)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().List(ctx, &FindSiloComments{
			EntryID:         testEntryID,
			IncludeResolved: true,
			IncludeDeleted:  true,
			FileID:          "file-1",
		})
		require.NoError(t, err)

		q := rec.last().URL.Query()
		assert.Equal(t, "true", q.Get("include_resolved"))
		assert.Equal(t, "true", q.Get("include_deleted"))
		assert.Equal(t, "file-1", q.Get("file_id"))
	})

	t.Run("fetch", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, testCommentRespond)
		c := testClient(t, rec)

		m, err := c.Silo().Comments().Fetch(ctx, testEntryID, testCommentID)
		require.NoError(t, err)
		assert.Equal(t, "Check the total", m.Body)
		assert.Equal(t, &SiloActor{Kind: SiloActorUser, ID: "u1"}, m.Author)

		req := rec.last()
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, testCommentPath, req.URL.Path)
	})

	t.Run("create", func(t *testing.T) {
		rec := newRecorder(http.StatusCreated, testCommentRespond)
		c := testClient(t, rec)

		m, err := c.Silo().Comments().Create(ctx, &CreateSiloComment{
			EntryID:   testEntryID,
			Body:      "Check the total",
			Paths:     []string{"$.doc.totals.payable"},
			File:      &SiloCommentFileRef{ID: "file-1", Page: 2},
			Important: true,
			AgentID:   testAgentID,
		})
		require.NoError(t, err)
		assert.Equal(t, testCommentID, m.ID)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/silo/v1/entries/entry-1/comments", req.URL.Path)
		assert.JSONEq(t, `{
			"body":"Check the total",
			"paths":["$.doc.totals.payable"],
			"file":{"id":"file-1","page":2},
			"important":true,
			"agent_id":"agent-1"
		}`, rec.lastBody(), "entry id travels in the path, not the body")
	})

	t.Run("reply", func(t *testing.T) {
		rec := newRecorder(http.StatusCreated, `{}`)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().Create(ctx, &CreateSiloComment{
			EntryID: testEntryID,
			ReplyTo: testCommentID,
			Body:    "Fixed",
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"reply_to":"`+testCommentID+`","body":"Fixed"}`, rec.lastBody())
	})

	t.Run("update", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, testCommentRespond)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().Update(ctx, &UpdateSiloComment{
			EntryID: testEntryID,
			ID:      testCommentID,
			Body:    "Check the payable total",
		})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, testCommentPath, req.URL.Path)
		assert.JSONEq(t, `{"body":"Check the payable total"}`, rec.lastBody())
	})

	for _, action := range []string{"resolve", "reopen"} {
		t.Run(action, func(t *testing.T) {
			rec := newRecorder(http.StatusOK, testCommentRespond)
			c := testClient(t, rec)
			svc := c.Silo().Comments()

			req := &ActOnSiloComment{EntryID: testEntryID, ID: testCommentID, AgentID: testAgentID}
			var err error
			if action == "resolve" {
				_, err = svc.Resolve(ctx, req)
			} else {
				_, err = svc.Reopen(ctx, req)
			}
			require.NoError(t, err)

			r := rec.last()
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, testCommentPath+"/"+action, r.URL.Path)
			assert.JSONEq(t, `{"agent_id":"agent-1"}`, rec.lastBody())
		})
	}

	t.Run("delete", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, testCommentRespond)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().Delete(ctx, &ActOnSiloComment{EntryID: testEntryID, ID: testCommentID})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodDelete, req.Method)
		assert.Equal(t, testCommentPath, req.URL.Path)
		assert.Empty(t, req.URL.RawQuery)
	})

	t.Run("delete as an agent", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, testCommentRespond)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().Delete(ctx, &ActOnSiloComment{EntryID: testEntryID, ID: testCommentID, AgentID: testAgentID})
		require.NoError(t, err)
		assert.Equal(t, testAgentID, rec.last().URL.Query().Get("agent_id"))
	})

	t.Run("full entry", func(t *testing.T) {
		rec := newRecorder(http.StatusUnprocessableEntity, `{"code":"resource_exhausted","message":"entry has too many comments"}`)
		c := testClient(t, rec)

		_, err := c.Silo().Comments().Create(ctx, &CreateSiloComment{EntryID: testEntryID, Body: "x"})
		require.Error(t, err)
		assert.True(t, IsUnprocessableEntity(err))
	})
}
