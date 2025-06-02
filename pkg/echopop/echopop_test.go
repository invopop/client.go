package echopop_test

import (
	"testing"

	"github.com/invopop/client.go/pkg/echopop"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetEnrollment(t *testing.T) {
	t.Run("nil key", func(t *testing.T) {
		c := echo.New().NewContext(nil, nil)
		assert.NotPanics(t, func() {
			en := echopop.GetEnrollment(c)
			assert.Nil(t, en)
		})
	})
}

func TestGetClient(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		c := echo.New().NewContext(nil, nil)
		assert.NotPanics(t, func() {
			ic := echopop.GetClient(c)
			assert.Nil(t, ic)
		})
	})
}
