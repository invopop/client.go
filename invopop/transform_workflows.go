package invopop

import (
	"context"
	"path"
)

const workflowsPath = "workflows"

// WorkflowsService encapsulates the functionality around workflows.
type WorkflowsService service

// CreateWorkflow sends a request to the API to create a new Workflow.
func (svc *WorkflowsService) Create(ctx context.Context, m *Workflow) error {
	p := path.Join(transformBasePath, workflowsPath, m.ID)
	return svc.client.put(ctx, p, m, m)
}

// List prepares a pageable list of workflows that belong to the requester.
func (svc *WorkflowsService) List(ctx context.Context, col *WorkflowCollection) error {
	return svc.client.get(ctx, path.Join(transformBasePath, workflowsPath), col)
}
