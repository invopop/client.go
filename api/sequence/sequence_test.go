package sequence_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/invopop/client/api/sequence"
	"github.com/stretchr/testify/assert"
)

const (
	ownerID = "fb3ad823-65fe-4a15-beff-7cfa085c7b29"
)

func TestCode(t *testing.T) {
	url := os.Getenv("INVOPOP_API_URL")
	if url == "" {
		t.Logf("run `export INVOPOP_API_URL=https://api.invopop.dev/sequence` to test")
		return
	}

	s := sequence.New(url)

	ctx := context.Background()

	nc, err := s.CreateCode(ctx, ownerID, &sequence.CodeParameters{
		Name:    "test",
		Prefix:  "test",
		Suffix:  "test",
		Padding: 5,
	})

	assert.Nil(t, err, "expecting nil error")

	cs, err := s.FetchCodeCollection(ctx, nc.Owner.Id)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, cs, "expecting non-nil codes")

	assert.Greater(t, len(cs.Codes), 0, "at least one code found")

	c, err := s.FetchCode(ctx, nc.Owner.Id, nc.Id)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, c, "expecting non-nil code")

	assert.Equal(t, c.Id, nc.Id, "fetch code same as created")
}

func TestFetchCodeError(t *testing.T) {
	url := os.Getenv("INVOPOP_API_URL")
	if url == "" {
		t.Logf("run `export INVOPOP_API_URL=https://api.invopop.dev/sequence` to test")
		return
	}

	s := sequence.New(url)

	ctx := context.Background()
	res, err := s.FetchCode(ctx, uuid.New().String(), uuid.New().String())

	assert.NotNil(t, err, "expecting error")
	assert.Nil(t, res, "expecting nil result")

	assert.Equal(t, http.StatusNotFound, err.Status, "no codes found")
}
