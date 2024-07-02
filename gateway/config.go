package gateway

import (
	"fmt"

	"github.com/invopop/configure"
	"github.com/invopop/configure/pkg/natsconf"
	"github.com/invopop/configure/pkg/zeroconf"
)

// Config holds the configuration for the gateway service.
type Config struct {
	// Service name registered with the gateway
	Name string `json:"name"`

	// Log defines a custom log configuration for the gateway.
	Log *zeroconf.Log `json:"log"`

	// WorkerCount
	WorkerCount int `json:"worker_count"`

	// NATS configuration used to connect to the gateway.
	NATS *natsconf.Config `json:"nats"`

	// Silo configuration options mainly for file upload.
	Silo *Silo `json:"silo"`
}

// Configuration defines what we expect from the config so that it
// can be overwritten if needed.
type Configuration interface {
	Init() error     // Prepare
	config() *Config // returns the original configuration internally
}

// Silo is used for uploading assets. We need a config to correctly configure
// where assets are uploaded to.
type Silo struct {
	PublicBaseURL string `json:"public_base_url"`
}

func (c *Config) config() *Config {
	return c
}

// Init prepares the configuration logs. If you want to extend
// this with your own implementation, be sure to also call this
// method.
func (c *Config) Init() error {
	c.Log.Init(c.Name)
	return nil
}

// ParseConfig attempts to load the file an populate the configuration
// object provided.
func ParseConfig(file string, conf Configuration) error {
	if err := configure.Load(file, conf); err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}
	return conf.Init()
}
