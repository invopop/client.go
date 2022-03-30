package invopop

import (
	"context"
	"path"
)

const (
	utilsBaseURL = "/utils/v1"
	pingPath     = "ping"
)

// Ping holds the simple ping request details.
type Ping struct {
	Ping string `json:"ping"`
}

// UtilsService is a simple implementation of the Invopop Utils API end point which
// can be used for connection and authentication testing.
type UtilsService service

// Ping updates the provided Ping instance with the payload from the server.
func (svc *UtilsService) Ping(ctx context.Context, p *Ping) error {
	return svc.client.get(ctx, path.Join(utilsBaseURL, pingPath), p)
}
