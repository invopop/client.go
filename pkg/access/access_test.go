package access

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/invopop/client.go/invopop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const syncResponse = `{"app_id":"myapp","actions":[{"id":"myapp.send","status":"created","rev":"1-a"}],"undeclared":["myapp.old"]}`

func testRequest() *invopop.SyncActions {
	return &invopop.SyncActions{Actions: []*invopop.ActionDefinition{{ID: "myapp.send", Name: "Send"}}}
}

// testServer replies to each request with the next of statuses, and with the
// last of them once they run out, counting the requests.
func testServer(t *testing.T, statuses ...int) (*invopop.Client, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		assert.Equal(t, "/access/v1/app/actions/sync", r.URL.Path)
		status := statuses[min(n, len(statuses))-1]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(syncResponse))
			return
		}
		_, _ = w.Write([]byte(`{"code":"error","message":"failed"}`))
	}))
	t.Cleanup(srv.Close)
	ic := invopop.New(invopop.WithConfig(&invopop.Config{
		BaseURL:      srv.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}))
	return ic, calls
}

func TestSyncActions(t *testing.T) {
	ctx := context.Background()
	fast := WithBackoff(time.Millisecond)

	t.Run("returns the result", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusOK)
		res, err := SyncActions(ctx, ic, testRequest(), fast)
		require.NoError(t, err)
		assert.Equal(t, "myapp", res.AppID)
		assert.Equal(t, []string{"myapp.old"}, res.Undeclared)
		assert.EqualValues(t, 1, calls.Load())
	})

	t.Run("retries server errors", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusServiceUnavailable, http.StatusTooManyRequests, http.StatusOK)
		_, err := SyncActions(ctx, ic, testRequest(), fast)
		require.NoError(t, err)
		assert.EqualValues(t, 3, calls.Load())
	})

	t.Run("gives up after the attempts", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusBadGateway)
		_, err := SyncActions(ctx, ic, testRequest(), fast, WithAttempts(3))
		require.Error(t, err)
		assert.EqualValues(t, 3, calls.Load())
	})

	t.Run("doesn't retry refused definitions", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusUnprocessableEntity)
		_, err := SyncActions(ctx, ic, testRequest(), fast)
		require.Error(t, err)
		assert.EqualValues(t, 1, calls.Load())
	})

	t.Run("retries when Access can't be reached", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		srv.Close()
		ic := invopop.New(invopop.WithConfig(&invopop.Config{BaseURL: srv.URL, ClientID: "id", ClientSecret: "secret"}))
		start := time.Now()
		_, err := SyncActions(ctx, ic, testRequest(), WithBackoff(10*time.Millisecond), WithAttempts(3))
		require.Error(t, err)
		// Two waits, of 10ms and 20ms.
		assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond)
	})

	t.Run("doesn't retry requests it can't make", func(t *testing.T) {
		ic := invopop.New()
		_, err := SyncActions(ctx, ic, testRequest(), fast)
		assert.ErrorIs(t, err, invopop.ErrMissingCredentials)
	})

	t.Run("doesn't retry definitions it can't encode", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusOK)
		req := testRequest()
		req.Actions[0].ConfigUI = []byte(`{not json`)
		// A retry would wait until the context ends.
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		_, err := SyncActions(ctx, ic, req, WithBackoff(time.Hour))
		require.Error(t, err)
		assert.NotErrorIs(t, err, context.DeadlineExceeded)
		assert.Zero(t, calls.Load())
	})

	t.Run("retries attempts that time out", func(t *testing.T) {
		calls := new(atomic.Int32)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) == 1 {
				// Slower than the attempt's timeout.
				select {
				case <-r.Context().Done():
				case <-time.After(time.Second):
				}
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(syncResponse))
		}))
		defer srv.Close()
		ic := invopop.New(invopop.WithConfig(&invopop.Config{BaseURL: srv.URL, ClientID: "id", ClientSecret: "secret"}))
		_, err := SyncActions(ctx, ic, testRequest(), fast, WithTimeout(50*time.Millisecond))
		require.NoError(t, err)
		assert.EqualValues(t, 2, calls.Load())
	})

	t.Run("stops when the context ends", func(t *testing.T) {
		ic, calls := testServer(t, http.StatusServiceUnavailable)
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		_, err := SyncActions(ctx, ic, testRequest(), WithBackoff(time.Hour))
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.EqualValues(t, 1, calls.Load())
	})
}
