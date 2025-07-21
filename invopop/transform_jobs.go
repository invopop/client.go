package invopop

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
)

const (
	jobsPath    = "jobs"
	intentsPath = "intents"
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
	Args map[string]string `json:"args,omitempty" title:"Args" description:"Any additional arguments that might be relevant for processing."`
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
	Fields   string `json:"fields,omitempty" title:"Fields" description:"Nested validation field errors"`
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
	Index   int32             `json:"index"`
	Status  string            `json:"status"`
	At      string            `json:"at"`
	Code    string            `json:"code,omitempty"`
	Args    map[string]string `json:"args,omitempty"`
	Message string            `json:"message,omitempty"`
}

// CreateJob is used to create new jobs
type CreateJob struct {
	ID         string `json:"-"`
	WorkflowID string `json:"workflow_id" form:"workflow_id" title:"Workflow ID" description:"WorkflowID description"`

	// Either Silo Entry ID, Data (complete envelope or document), Key and Args are
	// required in order to create a job.
	// If any combination are provided, Silo Entry ID will take priority, followed by data.
	SiloEntryID string            `json:"silo_entry_id,omitempty" form:"silo_entry_id" title:"Silo Entry ID" description:"ID for the entry in the silo as an alternative for the raw data object."`
	Data        json.RawMessage   `json:"data,omitempty" form:"data" title:"Data" description:"Raw JSON data of the GOBL Envelope or Object when the Silo Entry ID is empty."`
	Key         string            `json:"key,omitempty" form:"key" title:"Key" description:"Idempotency key to ensure that only one job will be created with this value."`
	Args        map[string]string `json:"args,omitempty" form:"args" title:"Arguments" description:"Additional arguments to associate with the job and may be used by actions."`

	Tags []string `json:"tags,omitempty" form:"tags" title:"Tags" description:"Tags to associate with the job."`

	// Time in seconds to block the connection waiting for a response on the server side.
	Wait int32 `json:"-"`
}

// UpdateIntent is used to issue new events for a Job's Intent while processing.
type UpdateIntent struct {
	ID      string `json:"id,omitempty" title:"ID" description:"UUID of the intent to update." example:"186522a6-e697-4e34-8498-eee961bcb845"`
	JobID   string `json:"job_id,omitempty" title:"Job ID" description:"UUID of the job to update." example:"186522a6-e697-4e34-8498-eee961bcb845"`
	Ref     string `json:"ref,omitempty" title:"Ref" description:"Reference code used to identify the intent when the id and job_id are not available."`
	Status  string `json:"status" title:"Status" description:"Status code of the event to add to the intent" example:"POKE"`
	Code    string `json:"code,omitempty" title:"Code" description:"Code of the event to add to the intent" example:"XX123"`
	Message string `json:"message,omitempty" title:"Message to include alongside the new event."`
}

// Create sends a request to the API to process a job. The `WithWait` request option can
// be used to have the server wait for a job to be completed before responding.
func (svc *JobsService) Create(ctx context.Context, req *CreateJob) (*Job, error) {
	p := path.Join(transformBasePath, jobsPath, req.ID)
	if req.Wait > 0 {
		p = fmt.Sprintf("%s?wait=%d", p, req.Wait)
	}
	m := new(Job)
	if req.ID != "" {
		return m, svc.client.put(ctx, p, req, m)
	}
	return m, svc.client.post(ctx, p, req, m)
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

// UpdateIntent is a special endpoint only usable by enrolled applications to update the status
// of a Job's intent during processing. Typcially this is used to poke an intent that is queued.
// This can only currently be used enrolled applications.
func (svc *JobsService) UpdateIntent(ctx context.Context, req *UpdateIntent) (*JobIntent, error) {
	p := path.Join(transformBasePath, jobsPath, intentsPath)
	m := new(JobIntent)
	return m, svc.client.post(ctx, p, req, m)
}

// PokeByRef is a convenience method that will build an UpdateIntent request for a simple
// POKE status.
func (svc *JobsService) PokeByRef(ctx context.Context, ref string) (*JobIntent, error) {
	req := &UpdateIntent{
		Ref:    ref,
		Status: "POKE",
	}
	return svc.UpdateIntent(ctx, req)
}
