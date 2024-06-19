package invopop

import (
	"encoding/json"
	"fmt"
)

const (
	transformBasePath = "/transform/v1"
)

// Job is responsible for executing a workflow on a specific GOBL envelope.
type Job struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	SiloEntryID string `json:"silo_entry_id,omitempty"`
	WorkflowID  string `json:"workflow_id"`

	Tags []string `json:"tags,omitempty"`

	CompletedAt string `json:"completed_at,omitempty"`

	Intents []*JobIntent `json:"intents,omitempty"`

	// Properties returned in response after completion
	Envelope    json.RawMessage `json:"envelope,omitempty"`
	Attachments []*Attachment   `json:"attachments,omitempty"`
}

// Status returns true if the job has completed, and if there were any problems
// executing the jobs, an error.
func (j *Job) Status() (bool, error) {
	if j.CompletedAt == "" {
		return false, nil
	}
	intent := j.Intents[len(j.Intents)-1]
	event := intent.Events[len(intent.Events)-1]
	if event.Status == "KO" {
		return true, fmt.Errorf("step %s failed at %s: %s", intent.StepID, event.At, event.Message)
	}
	return true, nil
}

// JobIntent represents an attempt to execute a task.
type JobIntent struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	StepID   string `json:"step_id" title:"Step ID" description:"ID of the step to use" example:"8d49556b-ff63-477b-9cd3-32c986c1c77b"`
	Name     string `json:"name" title:"Name" description:"Name of the executed workflow step" example:"PDF Generation"`
	Provider string `json:"provider" title:"Provider" description:"ID of the provider to use" example:"pdf"`

	Events []*JobIntentEvent `json:"events,omitempty"`

	Completed bool `json:"completed,omitempty"`
}

// JobIntentEvent represents the state and history of executing an intent.
type JobIntentEvent struct {
	Index   int32  `json:"index"`
	Status  string `json:"status"`
	At      string `json:"at"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

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
