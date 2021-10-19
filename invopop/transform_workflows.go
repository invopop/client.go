package invopop

import (
	"context"
	"fmt"
)

// CreateWorkflow sends a request to the API to create a new Workflow.
func (svc *TransformService) CreateWorkflow(ctx context.Context, m *Workflow) error {
	path := fmt.Sprintf("%s/workflows/%s", transformBasePath, m.ID)
	return svc.client.put(ctx, path, m)
}

// ListWorkflows prepares a pageable list of workflows that belong to the requester.
func (svc *TransformService) ListWorkflows(ctx context.Context, col *WorkflowCollection) error {
	return svc.client.get(ctx, transformBasePath+"/workflows", col)
}
