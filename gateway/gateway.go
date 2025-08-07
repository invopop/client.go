// Package gateway provides a wrapper around the Invopop gateway service used
// to respond to incoming messages to process tasks.
//
// This package is only meant to be used by applications that will receive and
// process tasks via NATS, hence why it is independent from the main Invopop
// package.
//
// Usage of zerolog for logging is assumed.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/invopop/configure/pkg/natsconf"
	nats "github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

const (
	defaultWorkerCount = 8
	defaultTaskTimeout = 1 * time.Minute
)

// Client wraps around the functionality provided by the
// the gateway service, accessed via NATS.
type Client struct {
	name              string // service name
	nc                *nats.Conn
	wg                sync.WaitGroup
	th                TaskHandler
	timeout           time.Duration
	incoming          chan *nats.Msg
	sub               *nats.Subscription
	siloPublicBaseURL string
	workerCount       int
}

// Option provides a way to configure the gateway client using a
// set of functions.
type Option func(gw *Client)

// New instantiates a new gateway client service using the provided configuration
// options. Previous versions expected a configuration directly, you can
// still use that method if you prefer as follows:
//
//		gw := gateway.New(
//	 	gateway.WithConfig(conf),
//	 	gateway.WithTaskHandler(handler),
//	 )
//
// If you already have a nats connection you can use that directly, but
// be sure to set the additional name options:
//
//	gw := gateway.New(
//		gateway.WithName("my-service"),
//		gateway.WithNATS(nc),
//		gateway.WithTaskHandler(handler),
//	)
func New(opts ...Option) *Client {
	gw := new(Client)
	gw.incoming = make(chan *nats.Msg)

	for _, opt := range opts {
		opt(gw)
	}

	if gw.workerCount == 0 {
		gw.workerCount = defaultWorkerCount
	}
	if gw.timeout == 0 {
		gw.timeout = defaultTaskTimeout
	}

	return gw
}

// NATS provides the NATS Connection so that it can be used
// elsewhere if needed.
func (gw *Client) NATS() *nats.Conn {
	return gw.nc
}

// Subscribe indicates which method should be called when
// messages are received from the gateway service.
//
// Deprecated: Use the WithTaskHandler option during instantiation
// instead.
func (gw *Client) Subscribe(th TaskHandler) {
	gw.th = th
}

// Start begins the gateway service and starts listening for incoming tasks.
func (gw *Client) Start() error {
	if gw.name == "" {
		return errors.New("name required")
	}
	if gw.th == nil {
		return errors.New("task handler required")
	}
	if gw.nc == nil {
		return errors.New("nats connection required")
	}
	if err := gw.subscribeIncomingTasks(); err != nil {
		return fmt.Errorf("subscribing for tasks: %w", err)
	}
	log.Debug().Int("count", gw.workerCount).Msg("gateway: starting workers")
	for i := 0; i < gw.workerCount; i++ {
		go gw.startTaskWorker()
	}
	return nil
}

// Stop is used to gracefully drain all requests and wait for them to complete.
func (gw *Client) Stop() {
	tn := time.Now()
	log.Debug().Msg("gateway: shutting down")

	if gw.sub != nil {
		gw.sub.Unsubscribe() // nolint:errcheck
		gw.sub.Drain()       // nolint:errcheck
	}
	close(gw.incoming) // this stops workers from receiving more
	gw.wg.Wait()

	log.Info().Dur("dur", time.Since(tn)).Msg("gateway: shutdown complete")
}

func (gw *Client) subscribeIncomingTasks() error {
	subj := fmt.Sprintf(SubjectTaskFmt, gw.name)
	queue := fmt.Sprintf(QueueNameTaskFmt, gw.name)
	var err error
	gw.sub, err = gw.nc.QueueSubscribeSyncWithChan(subj, queue, gw.incoming)
	if err != nil {
		return fmt.Errorf("error subscribing to queue: %w", err)
	}
	return nil
}

func (gw *Client) startTaskWorker() {
	for m := range gw.incoming {
		gw.processTask(m)
	}
}

func (gw *Client) processTask(m *nats.Msg) {
	gw.wg.Add(1)
	defer gw.wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), gw.timeout)
	defer cancel()

	// Handling the incoming data
	t := new(Task)
	var res *TaskResult
	if err := proto.Unmarshal(m.Data, t); err != nil {
		res = TaskError(fmt.Errorf("parsing incoming task: %w", err))
	} else {
		// Handle panics from task handler
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Get stack trace for debugging
					stack := debug.Stack()

					// Log the panic with full details for monitoring
					log.Error().
						Str("task_id", t.Id).
						Str("action", t.Action).
						Str("trace", string(stack)).
						Str("job_id", t.JobId).
						Str("owner_id", t.OwnerId).
						Str("silo_entry_id", t.SiloEntryId).
						Msgf("[PANIC RECOVERED] %v", r)

					// Convert panic to user-friendly TaskKO so that we stop any
					// future retries. We assume here that retrying will not work
					// until the underlying issue is fixed.
					res = TaskKO(fmt.Errorf("unexpected data error"))
				}
			}()

			res = gw.th(ctx, t)
			if res == nil {
				// assume the response is okay if no content
				res = TaskOK()
			}
		}()
	}

	// Send the reply back
	data, err := proto.Marshal(res)
	if err != nil {
		log.Error().Str("task_id", t.Id).Err(err).Msg("unable to marshal task response, dropping")
	}
	if err := gw.nc.Publish(m.Reply, data); err != nil {
		log.Error().Str("task_id", t.Id).Err(err).Msg("unable to publish response")
	}
}

func prepareNATSClient(conf *natsconf.Config, name string) *nats.Conn {
	// prepare base options
	opts, err := conf.Options()
	if err != nil {
		log.Fatal().Err(err).Msg("preparing nats options")
	}

	opts = append(opts,
		nats.Name(name),
		// Add our own connection logging stuff
		nats.ConnectHandler(func(_ *nats.Conn) {
			log.Info().Str("url", conf.URL).Msg("nats connected")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Warn().Str("url", conf.URL).Err(err).Msg("nats disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Info().Str("url", conf.URL).Msg("nats reconnected")
		}),
		nats.ReconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Warn().Str("url", conf.URL).Err(err).Msg("nats reconnect error")
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			log.Warn().Str("url", conf.URL).Msg("nats closed")
		}),
	)

	// Create the connection
	nc, err := nats.Connect(conf.URL, opts...)
	if err != nil {
		log.Fatal().Err(err).Str("url", conf.URL).Msg("failed to connect to nats")
	}

	return nc
}
