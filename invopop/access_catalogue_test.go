package invopop

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSyncCatalogue() *SyncCatalogue {
	return &SyncCatalogue{
		Operations: []*OperationDefinition{
			{
				Key:     "invoice",
				Title:   "Invoice",
				Effect:  OperationEffectRead,
				Path:    "/entries/{entry_id}/invoice",
				Example: json.RawMessage(`{"lines":[]}`),
				Scope:   "read:silo:entries",
				Schemas: []string{testSchema},
			},
			{
				Key:         "set-purchase-lines",
				Title:       "Set a purchase's accounts",
				Description: "Sets the account of each line of a purchase.",
				Effect:      OperationEffectWrite,
				Method:      http.MethodPatch,
				Path:        "/purchases/{id}/lines",
				Params:      []*OperationParam{{Name: "id", Description: "The purchase's ID."}},
				BodySchema:  json.RawMessage(`{"type":"object"}`),
				Scope:       "admin:silo:entries",
			},
		},
	}
}

func TestCatalogueSync(t *testing.T) {
	const response = `{
		"app_id": "myapp",
		"operations": {"status": "updated", "digest": "1a010b7548f99f08"}
	}`

	t.Run("sends the operations with the app credentials", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Catalogue().Sync(context.Background(), testSyncCatalogue())
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/access/v1/app/catalogue/sync", req.URL.Path)
		id, secret, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		assert.Equal(t, "client-id", id)
		assert.Equal(t, "client-secret", secret)

		assert.JSONEq(t, `{
			"operations": [{
				"key": "invoice",
				"title": "Invoice",
				"effect": "read",
				"path": "/entries/{entry_id}/invoice",
				"example": {"lines": []},
				"scope": "read:silo:entries",
				"schemas": ["bill/invoice"]
			}, {
				"key": "set-purchase-lines",
				"title": "Set a purchase's accounts",
				"description": "Sets the account of each line of a purchase.",
				"effect": "write",
				"method": "PATCH",
				"path": "/purchases/{id}/lines",
				"params": [{"name": "id", "description": "The purchase's ID."}],
				"body_schema": {"type": "object"},
				"scope": "admin:silo:entries"
			}]
		}`, rec.lastBody())

		assert.Equal(t, "myapp", res.AppID)
		require.NotNil(t, res.Operations)
		assert.Equal(t, SyncedPartUpdated, res.Operations.Status)
		assert.Equal(t, "1a010b7548f99f08", res.Operations.Digest)
	})

	t.Run("uses the app credentials instead of an auth token", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec).WithAuthToken(testToken)

		_, err := c.Access().Catalogue().Sync(context.Background(), testSyncCatalogue())
		require.NoError(t, err)

		id, secret, ok := rec.last().BasicAuth()
		require.True(t, ok, "expected basic auth, got %q", rec.last().Header.Get("Authorization"))
		assert.Equal(t, "client-id", id)
		assert.Equal(t, "client-secret", secret)
	})

	t.Run("sends an empty list, which removes every operation", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec)

		_, err := c.Access().Catalogue().Sync(context.Background(), &SyncCatalogue{Operations: []*OperationDefinition{}})
		require.NoError(t, err)
		assert.JSONEq(t, `{"operations":[]}`, rec.lastBody())
	})

	t.Run("sends skills when there is a list of them", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"app_id": "myapp", "operations": {"status": "unchanged"},
			"skills": {"status": "created", "digest": "ffee0011aabb2233", "keys": ["myapp:purchase-accounts"]}}`)
		c := testOAuthClient(t, rec)

		req := &SyncCatalogue{
			Operations: []*OperationDefinition{},
			Skills: []*SkillDefinition{{
				Key:          "purchase-accounts",
				Title:        "Purchase accounts",
				Instructions: "Assign accounts.",
				Sources:      map[string]*SkillSource{"purchase": {Operation: "myapp:purchase-coding"}},
				Operations: []*SkillOperation{{
					Operation: "myapp:set-purchase-lines", Name: "set_accounts",
					Params: map[string]string{"id": "{sources.purchase.id}"},
					Fixed:  map[string]any{"set_by": "barn:{agent_id}"},
				}},
				Results: []*SkillResult{{Code: "assigned", Status: SkillResultOK}},
			}},
		}
		res, err := c.Access().Catalogue().Sync(context.Background(), req)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"operations": [],
			"skills": [{
				"key": "purchase-accounts",
				"title": "Purchase accounts",
				"instructions": "Assign accounts.",
				"sources": {"purchase": {"operation": "myapp:purchase-coding"}},
				"operations": [{
					"operation": "myapp:set-purchase-lines",
					"name": "set_accounts",
					"params": {"id": "{sources.purchase.id}"},
					"fixed": {"set_by": "barn:{agent_id}"}
				}],
				"results": [{"code": "assigned", "status": "ok"}]
			}]
		}`, rec.lastBody())
		require.NotNil(t, res.Skills)
		assert.Equal(t, []string{"myapp:purchase-accounts"}, res.Skills.Keys)

		// An empty list removes every skill; no list leaves them alone.
		_, err = c.Access().Catalogue().Sync(context.Background(), &SyncCatalogue{Operations: []*OperationDefinition{}, Skills: []*SkillDefinition{}})
		require.NoError(t, err)
		assert.JSONEq(t, `{"operations":[],"skills":[]}`, rec.lastBody())
	})

	t.Run("sends panels when there is a list of them", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"app_id": "myapp", "operations": {"status": "unchanged"},
			"panels": [{"key": "invoice-tab", "digest": "1a010b7548f99f08"}]}`)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Catalogue().Sync(context.Background(), &SyncCatalogue{
			Operations: []*OperationDefinition{},
			Panels: []*PanelDefinition{{
				Key: "invoice-tab", Title: "Engine", Slots: []string{"entry.tab"}, Schemas: []string{testSchema},
				Sources: map[string]string{"books": "myapp:invoice"}, UI: map[string]any{"components": []any{}},
			}},
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"operations": [], "panels": [{
			"key": "invoice-tab", "title": "Engine", "slots": ["entry.tab"], "schemas": ["bill/invoice"],
			"sources": {"books": "myapp:invoice"}, "ui": {"components": []}
		}]}`, rec.lastBody())
		require.Len(t, res.Panels, 1)
		assert.Equal(t, "1a010b7548f99f08", res.Panels[0].Digest)
	})

	t.Run("requires the app credentials", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testClient(t, rec)

		_, err := c.Access().Catalogue().Sync(context.Background(), testSyncCatalogue())
		assert.ErrorIs(t, err, ErrMissingCredentials)
		assert.Empty(t, rec.requests)
	})

	t.Run("requires a list of operations", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec)

		_, err := c.Access().Catalogue().Sync(context.Background(), nil)
		assert.ErrorIs(t, err, ErrMissingOperations)
		_, err = c.Access().Catalogue().Sync(context.Background(), new(SyncCatalogue))
		assert.ErrorIs(t, err, ErrMissingOperations)
		assert.Empty(t, rec.requests)
	})

	t.Run("returns API errors", func(t *testing.T) {
		rec := newRecorder(http.StatusUnprocessableEntity, `{"message":"operations[0] \"invoice\": scope is missing"}`)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Catalogue().Sync(context.Background(), testSyncCatalogue())
		assert.Nil(t, res)
		assert.ErrorContains(t, err, "scope is missing")
	})
}
