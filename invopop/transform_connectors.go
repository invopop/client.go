package invopop

import (
	"context"
	"fmt"
)

// CreateConnector sends a request to the API to create a new connector.
func (svc *TransformService) CreateConnector(ctx context.Context, m *Connector) error {
	path := fmt.Sprintf("%s/tasks/%s", transformBasePath, m.ID)
	return svc.client.put(ctx, path, m)
}

// ListConnectors prepares a pageable list of connectors that belong to the requester.
func (svc *TransformService) ListConnectors(ctx context.Context, col *ConnectorCollection) error {
	return svc.client.get(ctx, transformBasePath+"/tasks", col)
}
