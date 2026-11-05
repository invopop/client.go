// Package access helps applications keep what they declare to Invopop's
// Access service up to date, such as the actions they provide.
package access

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/invopop/client.go/invopop"
	"github.com/rs/zerolog/log"
)

const (
	defaultAttempts = 5
	defaultBackoff  = 2 * time.Second
	defaultTimeout  = 30 * time.Second
)

// Option changes how SyncActions retries.
type Option func(*options)

type options struct {
	attempts int
	backoff  time.Duration
	timeout  time.Duration
}

// WithAttempts sets how many times to try the sync, 5 by default.
func WithAttempts(n int) Option {
	return func(o *options) {
		o.attempts = max(n, 1)
	}
}

// WithBackoff sets the delay before the first retry, which doubles on each
// one after, 2 seconds by default.
func WithBackoff(d time.Duration) Option {
	return func(o *options) {
		o.backoff = d
	}
}

// WithTimeout sets how long each attempt may take, 30 seconds by default.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

// SyncActions sends the application's action definitions to Access, such as
// on startup, and logs each action's outcome and any actions the application
// no longer declares. While Access can't be reached or fails, the sync is
// retried with a growing delay, until it succeeds, ctx ends or the attempts
// run out. Errors that would be the same again, such as definitions Access
// refuses, are returned straight away.
//
// The client must be configured with the application's OAuth client ID and
// secret, as for invopop.ActionsService.Sync.
func SyncActions(ctx context.Context, ic *invopop.Client, req *invopop.SyncActions, opts ...Option) (*invopop.SyncActionsResult, error) {
	o := &options{attempts: defaultAttempts, backoff: defaultBackoff, timeout: defaultTimeout}
	for _, opt := range opts {
		opt(o)
	}
	wait := o.backoff
	for attempt := 1; ; attempt++ {
		res, err := syncActions(ctx, ic, req, o.timeout)
		if err == nil {
			logResult(res)
			return res, nil
		}
		if attempt >= o.attempts || !retryable(err) || ctx.Err() != nil {
			return nil, err
		}
		log.Warn().Err(err).Int("attempt", attempt).Dur("retry_in", wait).Msg("access: syncing actions")
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		wait *= 2
	}
}

func syncActions(ctx context.Context, ic *invopop.Client, req *invopop.SyncActions, timeout time.Duration) (*invopop.SyncActionsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return ic.Access().Actions().Sync(ctx, req)
}

// retryable reports whether an error may not happen again: one from Access
// itself, too many requests, or no response at all because the connection
// failed or the attempt timed out. Anything else, such as definitions that
// can't be encoded or a request without credentials, fails the same way each
// time.
func retryable(err error) bool {
	var re *invopop.ResponseError
	if errors.As(err, &re) {
		code := re.StatusCode()
		return code >= http.StatusInternalServerError || code == http.StatusTooManyRequests
	}
	var ne net.Error
	return errors.As(err, &ne) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}

func logResult(res *invopop.SyncActionsResult) {
	for _, a := range res.Actions {
		log.Info().Str("app_id", res.AppID).Str("action", a.ID).Str("status", a.Status).Msg("access: synced action")
	}
	if len(res.Undeclared) > 0 {
		log.Warn().Str("app_id", res.AppID).Strs("actions", res.Undeclared).Msg("access: actions no longer declared")
	}
}
