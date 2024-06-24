package invopop

const (
	transformBasePath = "/transform/v1"
)

// TransformService provides access to the transform API end points.
type TransformService service

// Jobs provides the service to manage jobs.
func (svc *TransformService) Jobs() *JobsService {
	return (*JobsService)(svc)
}

// Workflows provides the service to manage workflows.
func (svc *TransformService) Workflows() *WorkflowsService {
	return (*WorkflowsService)(svc)
}
