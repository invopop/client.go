package gateway

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("applies defaults", func(t *testing.T) {
		gw := New()
		assert.Equal(t, defaultWorkerCount, gw.workerCount)
		assert.Equal(t, defaultTaskTimeout, gw.timeout)
		assert.NotNil(t, gw.incoming)
		assert.Nil(t, gw.NATS())
	})

	t.Run("applies options over defaults", func(t *testing.T) {
		gw := New(
			WithName("my-service"),
			WithWorkerCount(2),
			WithTaskTimeout(10*time.Second),
			WithSiloPublicBaseURL("https://silo.example.com"),
		)
		assert.Equal(t, "my-service", gw.name)
		assert.Equal(t, 2, gw.workerCount)
		assert.Equal(t, 10*time.Second, gw.timeout)
		assert.Equal(t, "https://silo.example.com", gw.siloPublicBaseURL)
	})

	t.Run("sets the task handler", func(t *testing.T) {
		called := false
		th := func(_ context.Context, _ *Task) *TaskResult {
			called = true
			return TaskOK()
		}
		gw := New(WithTaskHandler(th))
		require.NotNil(t, gw.th)

		res := gw.th(context.Background(), new(Task))
		assert.True(t, called)
		assert.Equal(t, TaskStatus_OK, res.Status)
	})
}

func TestSubscribe(t *testing.T) {
	gw := New()
	require.Nil(t, gw.th)

	gw.Subscribe(func(_ context.Context, _ *Task) *TaskResult {
		return TaskSkip("nope")
	})
	require.NotNil(t, gw.th)
	assert.Equal(t, TaskStatus_SKIP, gw.th(context.Background(), new(Task)).Status)
}

func TestStartValidation(t *testing.T) {
	th := func(_ context.Context, _ *Task) *TaskResult { return TaskOK() }

	tests := []struct {
		name string
		gw   *Client
		err  string
	}{
		{
			name: "missing name",
			gw:   New(WithTaskHandler(th)),
			err:  "name required",
		},
		{
			name: "missing task handler",
			gw:   New(WithName("svc")),
			err:  "task handler required",
		},
		{
			name: "missing nats connection",
			gw:   New(WithName("svc"), WithTaskHandler(th)),
			err:  "nats connection required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.gw.Start(), tt.err)
		})
	}
}
