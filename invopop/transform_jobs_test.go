package invopop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobDone(t *testing.T) {
	tests := []struct {
		name    string
		job     *Job
		done    bool
		wantErr string
	}{
		{
			name: "incomplete",
			job:  &Job{},
			done: false,
		},
		{
			name: "ok",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "OK",
				Intents: []*JobIntent{
					{StepID: "step-1", Events: []*JobIntentEvent{{Status: "OK"}}},
				},
			},
			done: true,
		},
		{
			name: "failed with event detail",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "KO",
				Intents: []*JobIntent{
					{StepID: "step-1", Events: []*JobIntentEvent{
						{Status: "KO", At: "2026-08-28T09:59:00.000Z", Message: "boom"},
					}},
				},
			},
			done:    true,
			wantErr: "step step-1 failed at 2026-08-28T09:59:00.000Z: boom",
		},
		{
			// The API omits `events` for an intent whose first event is not yet
			// visible; the job's faults still describe the failure.
			name: "failed with intent but no events falls back to faults",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "KO",
				Intents:     []*JobIntent{{StepID: "step-1"}},
				Faults:      []*Fault{{Provider: "pdf", Message: "render failed"}},
			},
			done:    true,
			wantErr: "step pdf failed: render failed",
		},
		{
			name: "failed with intent but no events or faults",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "KO",
				Intents:     []*JobIntent{{StepID: "step-1"}},
			},
			done:    true,
			wantErr: "job failed",
		},
		{
			name: "completed with no intents at all",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "KO",
			},
			done:    true,
			wantErr: "job failed",
		},
		{
			name: "success with no intents at all",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Status:      "OK",
			},
			done: true,
		},
		{
			// Servers that predate the job status field.
			name: "no status falls back to the last event",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Intents: []*JobIntent{
					{StepID: "step-1", Events: []*JobIntentEvent{
						{Status: "KO", At: "2026-08-28T09:59:00.000Z", Message: "boom"},
					}},
				},
			},
			done:    true,
			wantErr: "step step-1 failed at 2026-08-28T09:59:00.000Z: boom",
		},
		{
			name: "no status and no events",
			job: &Job{
				CompletedAt: "2026-08-28T10:00:00.000Z",
				Intents:     []*JobIntent{{StepID: "step-1"}},
			},
			done: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done, err := tt.job.Done()
			assert.Equal(t, tt.done, done)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}
