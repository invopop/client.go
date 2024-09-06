package gateway

import (
	"time"

	nats "github.com/nats-io/nats.go"
)

// WithName is used to set the name of the provider.
func WithName(name string) Option {
	return func(gw *Client) {
		gw.name = name
	}
}

// WithConfig allows the gateway connection to be configured with a specific
// configuration object.
func WithConfig(conf Configuration) Option {
	return func(gw *Client) {
		gconf := conf.config()

		gw.name = gconf.Name
		gw.workerCount = gconf.WorkerCount
		gw.nc = prepareNATSClient(gconf.NATS, gconf.Name)

		if gconf.Silo != nil {
			gw.siloPublicBaseURL = gconf.Silo.PublicBaseURL
		}
	}
}

// WithTaskTimeout sets the amount of time to wait before cancelling a task.
// The default time is 1 minute.
func WithTaskTimeout(dur time.Duration) Option {
	return func(gw *Client) {
		gw.timeout = dur
	}
}

// WithTaskHandler configures where incoming tasks will be sent. Alternatively,
// the "Subscribe" method can be used to set the handler.
func WithTaskHandler(th TaskHandler) Option {
	return func(gw *Client) {
		gw.th = th
	}
}

// WithNATS configures the gateway  to use the provided
// NATS connection.
func WithNATS(nc *nats.Conn) Option {
	return func(gw *Client) {
		gw.nc = nc
	}
}

// WithWorkerCount sets the number of workers to use for processing.
func WithWorkerCount(count int) Option {
	return func(gw *Client) {
		gw.workerCount = count
	}
}

// WithSiloPublicBaseURL sets the public base URL for the silo which is used
// for uploading files. This is optional if you do not plan to upload files,
// or would rather use the API directly.
func WithSiloPublicBaseURL(url string) Option {
	return func(gw *Client) {
		gw.siloPublicBaseURL = url
	}
}
