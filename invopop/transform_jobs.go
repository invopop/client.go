package invopop

import (
	"context"
	"fmt"
)

// CreateJob sends a request to the API to process a job. The `WithWait` request option can
// be used to have the server wait for a job to be completed before responding.
func (svc *TransformService) CreateJob(ctx context.Context, m *Job, opts ...RequestOption) error {
	ro := handleOptions(opts)
	path := transformBasePath + "/jobs/" + m.ID
	if ro.wait > 0 {
		path = fmt.Sprintf("%s?wait=%d", path, ro.wait)
	}
	return svc.client.put(ctx, path, m)
}

// FetchJob fetches the latest job results. As with `CreateJob`, if the `WithWait` requestion
// option is defined, the server will wait for a completed job to be returned before timing out.
func (svc *TransformService) FetchJob(ctx context.Context, m *Job, opts ...RequestOption) error {
	ro := handleOptions(opts)
	path := transformBasePath + "/jobs/" + m.ID
	if ro.wait > 0 {
		path = fmt.Sprintf("%s?wait=%d", path, ro.wait)
	}
	return svc.client.get(ctx, path, m)
}
