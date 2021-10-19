package invopop

import "context"

const (
	pingBaseURL = "/ping/v1"
)

// Ping holds the simple ping response details.
type Ping struct {
	Ping string `json:"ping"`
}

// PingService is a simple implementation of the Invopop Ping API end point which
// can be used for connection and authentication testing.
type PingService service

// Fetch updates the provided Ping instance with the payload from the server.
func (svc *PingService) Fetch(ctx context.Context, p *Ping) error {
	return svc.client.get(ctx, pingBaseURL, p)
}
