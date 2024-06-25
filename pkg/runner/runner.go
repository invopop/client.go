// Package runner helps run applications using a set of defaults
// that make it easier to handle interrupts and shutdowns.
package runner

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	shutdownTimeout = 5 * time.Second
)

// RunFunc defines a function to run with a context.
type RunFunc func(ctx context.Context) error

// A Group keeps track of a set of things to run. Example:
//
//	g := new(Group)
//	g.Start(func(ctx context.Context) error {
//	  // start up services
//	})
//	g.Stop(func(ctx context.Context) error {
//	  // stop services
//	})
//	if err := g.Wait(); err != nil {
//	  // so something with error
//	}
type Group struct {
	start []RunFunc
	stop  []RunFunc
}

// Start adds a function to the group that will be called during startup.
func (g *Group) Start(f ...RunFunc) {
	g.start = append(g.start, f...)
}

// Stop adds a function to the group that will be called during shutdown.
func (g *Group) Stop(f ...RunFunc) {
	g.stop = append(g.stop, f...)
}

// Wait will first run all the start methods, then wait for a signal
// to shutdown, and finally run all the stop methods.
func (g *Group) Wait() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for _, f := range g.start {
		f := f
		go func() {
			if err := f(ctx); err != nil {
				panic(err)
			}
		}()
	}

	// Wait for a signal
	<-ctx.Done()

	return g.shutdown()
}

func (g *Group) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	sg := new(errgroup.Group)
	for _, f := range g.stop {
		f := f // prevent bug in loops
		sg.Go(func() error {
			return f(ctx)
		})
	}
	if err := sg.Wait(); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	// All good.
	return nil
}
