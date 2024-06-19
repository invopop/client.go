package invopop

import (
	"context"
	"fmt"
	"path"
)

const jobsPath = "jobs"

// JobsService provides endpoints for dealing with jobs.
type JobsService service

// CreateJob sends a request to the API to process a job. The `WithWait` request option can
// be used to have the server wait for a job to be completed before responding.
func (svc *JobsService) Create(ctx context.Context, m *Job, opts ...RequestOption) error {
	ro := handleOptions(opts)
	p := path.Join(transformBasePath, jobsPath, m.ID)
	if ro.wait > 0 {
		p = fmt.Sprintf("%s?wait=%d", p, ro.wait)
	}
	return svc.client.put(ctx, p, m, m)
}

// Fetch fetches the latest job results. As with `CreateJob`, if the `WithWait` requestion
// option is defined, the server will wait for a completed job to be returned before timing out.
func (svc *JobsService) Fetch(ctx context.Context, m *Job, opts ...RequestOption) error {
	ro := handleOptions(opts)
	p := path.Join(transformBasePath, jobsPath, m.ID)
	if ro.wait > 0 {
		p = fmt.Sprintf("%s?wait=%d", p, ro.wait)
	}
	return svc.client.get(ctx, p, m)
}
