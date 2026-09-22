package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupStart(t *testing.T) {
	g := new(Group)
	assert.Empty(t, g.start)

	noop := func(_ context.Context) error { return nil }
	g.Start(noop)
	assert.Len(t, g.start, 1)

	g.Start(noop, noop)
	assert.Len(t, g.start, 3, "should append rather than replace")
}

func TestGroupStop(t *testing.T) {
	g := new(Group)
	assert.Empty(t, g.stop)

	noop := func(_ context.Context) error { return nil }
	g.Stop(noop)
	assert.Len(t, g.stop, 1)

	g.Stop(noop, noop)
	assert.Len(t, g.stop, 3, "should append rather than replace")
}

func TestGroupShutdown(t *testing.T) {
	t.Run("with no stop functions", func(t *testing.T) {
		g := new(Group)
		assert.NoError(t, g.shutdown())
	})

	t.Run("runs every stop function", func(t *testing.T) {
		var count atomic.Int32
		g := new(Group)
		for range 3 {
			g.Stop(func(_ context.Context) error {
				count.Add(1)
				return nil
			})
		}
		require.NoError(t, g.shutdown())
		assert.Equal(t, int32(3), count.Load())
	})

	t.Run("wraps a stop error", func(t *testing.T) {
		g := new(Group)
		g.Stop(func(_ context.Context) error {
			return errors.New("could not close")
		})
		err := g.shutdown()
		assert.EqualError(t, err, "shutdown: could not close")
	})

	t.Run("runs the remaining stop functions when one fails", func(t *testing.T) {
		var count atomic.Int32
		g := new(Group)
		g.Stop(func(_ context.Context) error {
			return errors.New("first failed")
		})
		g.Stop(func(_ context.Context) error {
			count.Add(1)
			return nil
		})
		assert.Error(t, g.shutdown())
		assert.Equal(t, int32(1), count.Load())
	})

	t.Run("provides stop functions with a context that has a deadline", func(t *testing.T) {
		var deadline time.Time
		var ok bool
		g := new(Group)
		g.Stop(func(ctx context.Context) error {
			deadline, ok = ctx.Deadline()
			return nil
		})
		require.NoError(t, g.shutdown())
		require.True(t, ok, "shutdown context should have a deadline")
		assert.WithinDuration(t, time.Now().Add(shutdownTimeout), deadline, time.Second)
	})
}

func TestGroupWait(t *testing.T) {
	// Wait blocks until an interrupt arrives, so the start function raises
	// the signal itself. By the time start functions run, Wait has already
	// registered its signal handler.
	var started, stopped atomic.Bool

	g := new(Group)
	g.Start(func(_ context.Context) error {
		started.Store(true)
		return syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	})
	g.Stop(func(_ context.Context) error {
		stopped.Store(true)
		return nil
	})

	done := make(chan error, 1)
	go func() { done <- g.Wait() }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for shutdown")
	}

	assert.True(t, started.Load(), "start function should have run")
	assert.True(t, stopped.Load(), "stop function should have run")
}
