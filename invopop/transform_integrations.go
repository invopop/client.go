package invopop

import (
	"context"
	"path"
)

const integrationsPath = "integrations"

// CreateIntegration sends a request to the API to create a new integration.
func (svc *TransformService) CreateIntegration(ctx context.Context, m *Integration) error {
	p := path.Join(transformBasePath, integrationsPath, m.ID)
	return svc.client.put(ctx, p, m, m)
}

// ListIntegrations prepares a pageable list of integrations that belong to the requester.
func (svc *TransformService) ListIntegrations(ctx context.Context, col *IntegrationCollection) error {
	return svc.client.get(ctx, path.Join(transformBasePath, integrationsPath), col)
}
