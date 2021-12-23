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

	EnvelopeID string `json:"envelope_id,omitempty"`
	WorkflowID string `json:"workflow_id"`

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
	intent := j.Intents[len(j.Intents)]
	event := intent.Events[len(intent.Events)]
	if event.Status == "KO" {
		return true, fmt.Errorf("Task %s failed at %s: %s", intent.TaskID, event.At, event.Message)
	}
	return true, nil
}

// JobIntent represents an attempt to execute a task.
type JobIntent struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	TaskID string `json:"task_id"`

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

// Task defines a specific objective that we would like Invopop to be able to perform.
type Task struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Name  string `json:"name"`
	Notes string `json:"notes,omitempty"`

	Provider string          `json:"provider,omitempty"`
	Config   json.RawMessage `json:"config,omitempty"`

	Disabled bool `json:"disabled,omitempty"`
}

// TaskCollection contains a list of tasks.
type TaskCollection struct {
	List          []*Task `json:"list"`
	Limit         int32   `json:"limit"`
	CreatedAt     string  `json:"created_at,omitempty"`
	NextCreatedAt string  `json:"next_created_at,omitempty"`
}

// Workflow keeps together a list of tasks to execute when a job is requested.
type Workflow struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	TaskIDs []string `json:"task_ids,omitempty"`

	Disabled bool `json:"disabled,omitempty"`
}

// WorkflowCollection contains a list of workflows.
type WorkflowCollection struct {
	List      []*Workflow `json:"list"`
	Limit     int32       `json:"limit"`
	CreatedAt string      `json:"created_at"`
}

// TransformService provides access to the transform API end points.
type TransformService service
