package invopop

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
)

const (
	jobsPath    = "jobs"
	jobsKeyPath = "key"
)

// JobsService provides endpoints for dealing with jobs.
type JobsService service

// Job is responsible for executing a workflow on a specific GOBL envelope.
type Job struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	SiloEntryID string `json:"silo_entry_id,omitempty"`
	WorkflowID  string `json:"workflow_id"`

	Key  string            `json:"key,omitempty" title:"Key" description:"Key assigned to the job, used to identify it in the system."`
	Meta map[string]string `json:"meta,omitempty" title:"Meta" description:"Any additional data that might be relevant for processing."`
	Tags []string          `json:"tags,omitempty" title:"Tags" description:"Any tags that may be useful to be associated with the job."`

	CompletedAt string `json:"completed_at,omitempty"`

	Intents []*JobIntent `json:"intents,omitempty"`
	Faults  []*Fault     `json:"faults,omitempty" title:"Faults" description:"Array of fault objects that represent errors that occurred during the processing of the job."`

	// Properties returned in response after completion
	Envelope    json.RawMessage   `json:"envelope,omitempty"`
	Attachments []*SiloAttachment `json:"attachments,omitempty"`
}

// Fault represents an error that occurred during the processing of a job.
type Fault struct {
	Provider string `json:"provider" title:"Provider" description:"ID of the provider that generated the fault." example:"pdf"`
	Code     string `json:"code,omitempty" title:"Code" description:"Code assigned by the provider that may provide additional information about the fault."`
	Message  string `json:"message" title:"Message" description:"Message assigned by the provider that may provide additional information about the fault."`
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

// CreateJob is used to create new jobs
type CreateJob struct {
	ID         string `json:"-"`
	WorkflowID string `json:"workflow_id" form:"workflow_id" title:"Workflow ID" description:"WorkflowID description"`

	// Either Silo Entry ID, Data (complete envelope or document), Key and Meta are
	// required in order to create a job.
	// If any combination are provided, Silo Entry ID will take priority, followed by data.
	SiloEntryID string            `json:"silo_entry_id" form:"silo_entry_id" title:"Silo Entry ID" description:"ID for the entry in the silo as an alternative for the raw data object."`
	Data        json.RawMessage   `json:"data" form:"data" title:"Data" description:"Raw JSON data of the GOBL Envelope or Object when the Silo Entry ID is empty."`
	Key         string            `json:"key,omitempty" form:"key" title:"Key" description:"Idempotency key to ensure that only one job will be created with this value."`
	Meta        map[string]string `json:"meta,omitempty" form:"meta" title:"Meta" description:"Additional data to associate with the job."`

	Tags []string `json:"tags,omitempty" form:"tags" title:"Tags" description:"Tags to associate with the job."`

	// Time in seconds to block the connection waiting for a response on the server side.
	Wait int32 `json:"-"`
}

// Create sends a request to the API to process a job. The `WithWait` request option can
// be used to have the server wait for a job to be completed before responding.
func (svc *JobsService) Create(ctx context.Context, req *CreateJob) (*Job, error) {
	p := path.Join(transformBasePath, jobsPath, req.ID)
	if req.Wait > 0 {
		p = fmt.Sprintf("%s?wait=%d", p, req.Wait)
	}
	m := new(Job)
	return m, svc.client.put(ctx, p, req, m)
}

// Fetch fetches the latest job results.
func (svc *JobsService) Fetch(ctx context.Context, id string) (*Job, error) {
	p := path.Join(transformBasePath, jobsPath, id)
	m := new(Job)
	return m, svc.client.get(ctx, p, m)
}

// FetchByKey fetches the latest job results by its key
func (svc *JobsService) FetchByKey(ctx context.Context, key string) (*Job, error) {
	p := path.Join(transformBasePath, jobsPath, jobsKeyPath, key)
	m := new(Job)
	return m, svc.client.get(ctx, p, m)
}
