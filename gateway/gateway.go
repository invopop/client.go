// Package gateway provides a wrapper around the Invopop gateway service used
// to respond to incoming messages to process tasks.
//
// This package is only meant to be used by applications that will receive and
// process tasks via NATS, hence why it is independent from the main Invopop
// package.
package gateway
