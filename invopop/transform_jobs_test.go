package invopop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testCompletedAt = "2026-08-28T10:00:00.000Z"
	testStepID      = "step-1"
	testEventAt     = "2026-08-28T09:59:00.000Z"
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
				CompletedAt: testCompletedAt,
				Status:      "OK",
				Intents: []*JobIntent{
					{StepID: testStepID, Events: []*JobIntentEvent{{Status: "OK"}}},
				},
			},
			done: true,
		},
		{
			name: "failed with event detail",
			job: &Job{
				CompletedAt: testCompletedAt,
				Status:      "KO",
				Intents: []*JobIntent{
					{StepID: testStepID, Events: []*JobIntentEvent{
						{Status: "KO", At: testEventAt, Message: "boom"},
					}},
				},
			},
			done:    true,
			wantErr: "step step-1 failed at 2026-08-28T09:59:00.000Z: boom",
		},
		{
			name: "failed with intent but no events falls back to faults",
			job: &Job{
				CompletedAt: testCompletedAt,
				Status:      "KO",
				Intents:     []*JobIntent{{StepID: testStepID}},
				Faults:      []*Fault{{Provider: "pdf", Message: "render failed"}},
			},
			done:    true,
			wantErr: "step pdf failed: render failed",
		},
		{
			name: "failed with intent but no events or faults",
			job: &Job{
				CompletedAt: testCompletedAt,
				Status:      "KO",
				Intents:     []*JobIntent{{StepID: testStepID}},
			},
			done:    true,
			wantErr: "job failed",
		},
		{
			name: "completed with no intents at all",
			job: &Job{
				CompletedAt: testCompletedAt,
				Status:      "KO",
			},
			done:    true,
			wantErr: "job failed",
		},
		{
			name: "success with no intents at all",
			job: &Job{
				CompletedAt: testCompletedAt,
				Status:      "OK",
			},
			done: true,
		},
		{
			name: "no status falls back to the last event",
			job: &Job{
				CompletedAt: testCompletedAt,
				Intents: []*JobIntent{
					{StepID: testStepID, Events: []*JobIntentEvent{
						{Status: "KO", At: testEventAt, Message: "boom"},
					}},
				},
			},
			done:    true,
			wantErr: "step step-1 failed at 2026-08-28T09:59:00.000Z: boom",
		},
		{
			name: "no status and no events",
			job: &Job{
				CompletedAt: testCompletedAt,
				Intents:     []*JobIntent{{StepID: testStepID}},
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
