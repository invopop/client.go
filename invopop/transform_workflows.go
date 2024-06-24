package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"path"

	"github.com/invopop/gobl/uuid"
)

const workflowsPath = "workflows"

// WorkflowsService encapsulates the functionality around workflows.
type WorkflowsService service

// Workflow keeps together a list of integrations to execute when a job is requested.
type Workflow struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Name        string `json:"name" title:"Name" description:"Name of the workflow"`
	Description string `json:"description,omitempty" title:"Description" description:"Description of the workflow"`
	Schema      string `json:"schema,omitempty" title:"Schema" description:"Short schema name that the workflow will be allowed to process."`
	Country     string `json:"country,omitempty" title:"Country" description:"ISO country code the workflow will be used for."`
	Version     string `json:"version,omitempty" title:"Version" description:"Version of the workflow's contents currently defined."`

	Steps []*Step `json:"steps" title:"Steps" description:"List of steps to execute"`

	Disabled bool `json:"disabled,omitempty"`
}

// Step represents a single action inside a workflow
type Step struct {
	ID       string          `json:"id" title:"ID" description:"The UUID (any version) of the step." example:"186522a6-e697-4e34-8498-eee961bcb845"`
	Name     string          `json:"name" title:"Name" description:"Name of the step"`
	Provider string          `json:"provider" title:"Provider" description:"ID of the provider to use" example:"provider"`
	Notes    string          `json:"notes,omitempty" title:"Notes" description:"Additional internal details"`
	Config   json.RawMessage `json:"config,omitempty" title:"Configuration" description:"JSON configuration sent to the provider"`
	Next     []*Next         `json:"next,omitempty" title:"Next" description:"Optional array of next steps to execute after this one."`
}

// Next describes a next step to execute in a workflow.
type Next struct {
	Status string `json:"status,omitempty" title:"Status" description:"Step status to match against, when empty this next step will always be executed." enum:"OK,SKIP,KO,TIMEOUT"`
	StepID string `json:"step_id,omitempty" title:"Step ID" description:"ID of the step to execute next." example:"186522a6-e697-4e34-8498-eee961bcb845"`
	Stop   bool   `json:"stop,omitempty" title:"Stop" description:"When true, the workflow will stop after completing this step."`
}

// WorkflowCollection contains a list of workflows.
type WorkflowCollection struct {
	List      []*Workflow `json:"list"`
	Limit     int32       `json:"limit"`
	CreatedAt string      `json:"created_at"`
}

// CreateWorkflow defines what is required for a new workflow.
type CreateWorkflow struct {
	ID          string  `json:"-"`
	Name        string  `json:"name" form:"name" title:"Name" description:"Name of the workflow."`
	Description string  `json:"description,omitempty" form:"description" title:"Description" description:"Description of the workflow."`
	Schema      string  `json:"schema,omitempty" form:"schema" title:"Schema" description:"Short schema name that the workflow will be allowed to process." example:"bill/invoice"`
	Country     string  `json:"country,omitempty" form:"country" title:"Country Code" description:"ISO country code the workflow will be used in." example:"ES"`
	Steps       []*Step `json:"steps" form:"steps" title:"Steps" description:"Array of Steps to use for this workflow."`
}

// UpdateWorkflow defines what we can update in a workflow.
type UpdateWorkflow struct {
	ID          string  `json:"-"`
	Name        string  `json:"name" form:"name" title:"Name" description:"New name for the workflow."`
	Description string  `json:"description,omitempty" form:"description" title:"Description" description:"Updated description."`
	Steps       []*Step `json:"steps" form:"steps" title:"Steps" description:"Array of Steps to use for this workflow."`
}

// Fetch makes a request for the workflow by its ID.
func (sv *WorkflowsService) Fetch(ctx context.Context, id string) (*Workflow, error) {
	m := new(Workflow)
	return m, sv.client.get(ctx, path.Join(transformBasePath, workflowsPath, id), m)
}

// Create sends a request to the API to create a new Workflow.
func (svc *WorkflowsService) Create(ctx context.Context, req *CreateWorkflow) (*Workflow, error) {
	if req.ID == "" {
		req.ID = uuid.V7().String()
	}
	p := path.Join(transformBasePath, workflowsPath, req.ID)
	m := new(Workflow)
	return m, svc.client.put(ctx, p, req, m)
}

// Update will update the workflow.
func (svc *WorkflowsService) Update(ctx context.Context, req *UpdateWorkflow) (*Workflow, error) {
	if req.ID == "" {
		return nil, errors.New("missing workflow ID")
	}
	p := path.Join(transformBasePath, workflowsPath, req.ID)
	m := new(Workflow)
	return m, svc.client.patch(ctx, p, req, m)
}

// List prepares a pageable list of workflows that belong to the requester.
func (svc *WorkflowsService) List(ctx context.Context, createdAt string) (*WorkflowCollection, error) {
	p := path.Join(transformBasePath, workflowsPath)
	query := make(url.Values)
	if createdAt != "" {
		query.Add("created_at", createdAt)
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	m := new(WorkflowCollection)
	return m, svc.client.get(ctx, p, m)
}
