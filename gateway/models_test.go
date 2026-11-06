package gateway

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskResultHelpers(t *testing.T) {
	t.Run("task error", func(t *testing.T) {
		tr := TaskError(errors.New("something went wrong"))
		require.NotNil(t, tr)
		assert.Equal(t, TaskStatus_ERR, tr.Status)
		assert.Equal(t, "something went wrong", tr.Message)
		assert.Zero(t, tr.RetryIn)
	})

	t.Run("task ko", func(t *testing.T) {
		tr := TaskKO(errors.New("unrecoverable"))
		require.NotNil(t, tr)
		assert.Equal(t, TaskStatus_KO, tr.Status)
		assert.Equal(t, "unrecoverable", tr.Message)
	})

	t.Run("task ok", func(t *testing.T) {
		tr := TaskOK()
		require.NotNil(t, tr)
		assert.Equal(t, TaskStatus_OK, tr.Status)
		assert.Empty(t, tr.Message)
	})

	t.Run("task skip", func(t *testing.T) {
		tr := TaskSkip("nothing to do")
		require.NotNil(t, tr)
		assert.Equal(t, TaskStatus_SKIP, tr.Status)
		assert.Equal(t, "nothing to do", tr.Message)
	})

	t.Run("task queued", func(t *testing.T) {
		tr := TaskQueued("waiting for provider", 30)
		require.NotNil(t, tr)
		assert.Equal(t, TaskStatus_QUEUED, tr.Status)
		assert.Equal(t, "waiting for provider", tr.Message)
		assert.Equal(t, int32(30), tr.RetryIn)
	})

	t.Run("wraps a gateway error", func(t *testing.T) {
		// Gateway Errors are also errors, so they should serialize via Error().
		tr := TaskError(error(&Error{Code: ErrorCode_INVALID, Message: "bad data"}))
		assert.Equal(t, "INVALID: bad data", tr.Message)
	})
}

func TestTaskResultWithFaults(t *testing.T) {
	t.Run("chained on ko", func(t *testing.T) {
		tr := TaskKO(errors.New("validation failed")).WithFaults(
			&Fault{Code: "required", Message: "missing tax ID", Paths: []string{"$.supplier.tax_id"}},
			&Fault{Code: "format", Message: "bad date"},
		)
		assert.Equal(t, TaskStatus_KO, tr.Status)
		assert.Equal(t, "validation failed", tr.Message)
		require.Len(t, tr.Faults, 2)
		assert.Equal(t, []string{"$.supplier.tax_id"}, tr.Faults[0].Paths)
	})

	t.Run("appends on repeat calls", func(t *testing.T) {
		tr := TaskKO(errors.New("x")).WithFaults(&Fault{Code: "a"}).WithFaults(&Fault{Code: "b"})
		require.Len(t, tr.Faults, 2)
		assert.Equal(t, "b", tr.Faults[1].Code)
	})

	t.Run("none", func(t *testing.T) {
		tr := TaskKO(errors.New("x")).WithFaults()
		assert.Empty(t, tr.Faults)
	})
}
