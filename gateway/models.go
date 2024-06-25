package gateway

import "context"

// Subject and Queue names
const (
	SubjectTaskFmt     = "gw.%s.task" // for specific task messages
	SubjectFilesCreate = "gw.files.create"
	SubjectTasksPoke   = "gw.tasks.poke"
	SubjectStoreGet    = "gw.store.get"
	SubjectStoreSet    = "gw.store.set"
	QueueNameTaskFmt   = "%s.tasks"
)

// MIME Content Types supported by the "silo" service.
const (
	MIMEApplicationJSON           = "application/json"             // Complete Envelope
	MIMEApplicationMergePatchJSON = "application/merge-patch+json" // RFC7396
	MIMEApplicationJSONPatch      = "application/json-patch+json"  // RFC6902
)

// TaskHandler defines what type of method we need to call when
// an incoming task message is received.
type TaskHandler func(ctx context.Context, task *Task) *TaskResult

// TaskError is used to create a more friendly version of an error that can
// be sent back to the client.
func TaskError(err error) *TaskResult {
	return &TaskResult{
		Status:  TaskStatus_ERR,
		Message: err.Error(),
	}
}

// TaskKO is used to create a more friendly version of an error that cannot
// be recovered from without sending a new document. Something has clearly gone
// wrong in the configuration.
func TaskKO(err error) *TaskResult {
	return &TaskResult{
		Status:  TaskStatus_KO,
		Message: err.Error(),
	}
}

// TaskOK provides a simple task result indicating that everything went fine.
func TaskOK() *TaskResult {
	return &TaskResult{Status: TaskStatus_OK}
}

// TaskSkip provides skip response with the provided message. Skip is
// an informative response that implies the task was not executed for
// whatever reason, but we can safely continue processing.
func TaskSkip(msg string) *TaskResult {
	return &TaskResult{Status: TaskStatus_SKIP, Message: msg}
}
