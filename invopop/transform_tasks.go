package invopop

import (
	"context"
	"fmt"
)

// CreateTasks sends a request to the API to create a new task.
func (svc *TransformService) CreateTask(ctx context.Context, m *Task) error {
	path := fmt.Sprintf("%s/tasks/%s", transformBasePath, m.ID)
	return svc.client.put(ctx, path, m)
}

// ListTasks prepares a pageable list of tasks that belong to the requester.
func (svc *TransformService) ListTasks(ctx context.Context, col *TaskCollection) error {
	return svc.client.get(ctx, transformBasePath+"/tasks", col)
}
