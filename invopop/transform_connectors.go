package invopop

import (
	"context"
	"path"
)

const connectorsPath = "connectors"

// CreateConnector sends a request to the API to create a new connector.
func (svc *TransformService) CreateConnector(ctx context.Context, m *Connector) error {
	p := path.Join(transformBasePath, connectorsPath, m.ID)
	return svc.client.put(ctx, p, m)
}

// ListConnectors prepares a pageable list of connectors that belong to the requester.
func (svc *TransformService) ListConnectors(ctx context.Context, col *ConnectorCollection) error {
	return svc.client.get(ctx, path.Join(transformBasePath, connectorsPath), col)
}
