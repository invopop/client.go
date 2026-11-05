package invopop

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSyncActions() *SyncActions {
	return &SyncActions{
		Actions: []*ActionDefinition{
			{
				ID:           "myapp.send",
				Name:         "Send",
				ConfigUI:     json.RawMessage(`{"type":"panel"}`),
				Schemas:      []string{testSchema},
				Countries:    []string{"ES"},
				Results:      []*ActionResult{{Status: ActionResultOK, Code: "sent", Default: true}},
				Pops:         10,
				SeatRenewal:  SeatRenewalMonthly,
				ConfigSchema: json.RawMessage(`{"type":"object"}`),
			},
		},
	}
}

func TestActionsSync(t *testing.T) {
	const response = `{
		"app_id": "myapp",
		"actions": [{"id": "myapp.send", "status": "updated", "rev": "2-abc"}],
		"undeclared": ["myapp.old"]
	}`

	t.Run("sends the definitions with the app credentials", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Actions().Sync(context.Background(), testSyncActions())
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/access/v1/app/actions/sync", req.URL.Path)
		id, secret, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		assert.Equal(t, "client-id", id)
		assert.Equal(t, "client-secret", secret)

		assert.JSONEq(t, `{
			"actions": [{
				"id": "myapp.send",
				"name": "Send",
				"config_ui": {"type": "panel"},
				"config_schema": {"type": "object"},
				"schemas": ["bill/invoice"],
				"countries": ["ES"],
				"results": [{"status": "OK", "code": "sent", "default": true}],
				"pops": 10,
				"seat_renewal": "monthly"
			}]
		}`, rec.lastBody())

		assert.Equal(t, "myapp", res.AppID)
		require.Len(t, res.Actions, 1)
		assert.Equal(t, "myapp.send", res.Actions[0].ID)
		assert.Equal(t, SyncedActionUpdated, res.Actions[0].Status)
		assert.Equal(t, "2-abc", res.Actions[0].Rev)
		assert.Equal(t, []string{"myapp.old"}, res.Undeclared)
	})

	t.Run("uses the app credentials instead of an auth token", func(t *testing.T) {
		// Apps usually clone their client with an enrollment token; a bearer
		// header would replace the Basic one, and the API refuses tokens here.
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec).WithAuthToken(testToken)

		_, err := c.Access().Actions().Sync(context.Background(), testSyncActions())
		require.NoError(t, err)

		id, secret, ok := rec.last().BasicAuth()
		require.True(t, ok, "expected basic auth, got %q", rec.last().Header.Get("Authorization"))
		assert.Equal(t, "client-id", id)
		assert.Equal(t, "client-secret", secret)
	})

	t.Run("sends an empty list", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"app_id":"myapp","actions":[]}`)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Actions().Sync(context.Background(), &SyncActions{Actions: []*ActionDefinition{}})
		require.NoError(t, err)
		assert.JSONEq(t, `{"actions":[]}`, rec.lastBody())
		assert.Empty(t, res.Actions)
	})

	t.Run("requires the app credentials", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testClient(t, rec)

		_, err := c.Access().Actions().Sync(context.Background(), testSyncActions())
		assert.ErrorContains(t, err, "missing OAuth client credentials")
		assert.Empty(t, rec.requests)
	})

	t.Run("requires a list of actions", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, response)
		c := testOAuthClient(t, rec)

		_, err := c.Access().Actions().Sync(context.Background(), nil)
		assert.ErrorContains(t, err, "missing actions")
		_, err = c.Access().Actions().Sync(context.Background(), new(SyncActions))
		assert.ErrorContains(t, err, "missing actions")
		assert.Empty(t, rec.requests)
	})

	t.Run("returns API errors", func(t *testing.T) {
		rec := newRecorder(http.StatusForbidden, `{"message":"missing required scope: enrolled"}`)
		c := testOAuthClient(t, rec)

		res, err := c.Access().Actions().Sync(context.Background(), testSyncActions())
		assert.Nil(t, res)
		assert.True(t, IsForbidden(err), "expected a forbidden error, got %v", err)
	})
}
