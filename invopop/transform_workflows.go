package invopop

import (
	"context"
	"path"
)

const workflowsPath = "workflows"

// CreateWorkflow sends a request to the API to create a new Workflow.
func (svc *TransformService) CreateWorkflow(ctx context.Context, m *Workflow) error {
	p := path.Join(transformBasePath, workflowsPath, m.ID)
	return svc.client.put(ctx, p, m, m)
}

// ListWorkflows prepares a pageable list of workflows that belong to the requester.
func (svc *TransformService) ListWorkflows(ctx context.Context, col *WorkflowCollection) error {
	return svc.client.get(ctx, path.Join(transformBasePath, workflowsPath), col)
}
