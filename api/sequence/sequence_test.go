package sequence_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/invopop/client/api/sequence"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	url := os.Getenv("INVOPOP_API_URL")
	if url == "" {
		t.Logf("run `export INVOPOP_API_URL=https://api.invopop.dev` to test")
		return
	}

	apiKey := os.Getenv("INVOPOP_API_KEY")
	if apiKey == "" {
		t.Logf("run `export INVOPOP_API_KEY=<api_key>` to test")
		return
	}

	s := sequence.New(url, apiKey)

	runCode(t, s)
	runFetchCodeError(t, s)
	runEntry(t, s)
	runFetchEntryError(t, s)
}

func runCode(t *testing.T, s sequence.Client) {

	nc, err := s.CreateCode(&sequence.CodeParameters{
		ID:      uuid.New().String(),
		Name:    "test",
		Prefix:  "test",
		Suffix:  "test",
		Padding: 5,
	})

	assert.Nil(t, err, "expecting nil error")

	cs, err := s.FetchCodeCollection()

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, cs, "expecting non-nil codes")

	assert.Greater(t, len(cs.Codes), 0, "at least one code found")

	c, err := s.FetchCode(nc.ID)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, c, "expecting non-nil code")

	assert.Equal(t, c.ID, nc.ID, "fetch code same as created")
}

func runFetchCodeError(t *testing.T, s sequence.Client) {
	res, err := s.FetchCode(uuid.New().String())

	assert.NotNil(t, err, "expecting error")
	assert.Nil(t, res, "expecting nil result")
}

func runEntry(t *testing.T, s sequence.Client) {

	nc, err := s.CreateCode(&sequence.CodeParameters{
		ID:      uuid.New().String(),
		Name:    "test",
		Prefix:  "test",
		Suffix:  "test",
		Padding: 5,
	})

	assert.Nil(t, err, "expecting nil error")

	ne, err := s.CreateEntry(nc.ID, &sequence.EntryParameters{
		ID: uuid.New().String(),
		Meta: map[string]string{
			"key": "value",
		},
	})

	assert.Nil(t, err, "expecting nil error")

	es, err := s.FetchEntryCollection(nc.ID)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, es, "expecting non-nil entries")

	assert.Greater(t, len(es.Entries), 0, "at least one entry found")

	e, err := s.FetchEntry(nc.ID, ne.ID)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, e, "expecting non-nil entry")
}

func runFetchEntryError(t *testing.T, s sequence.Client) {
	res, err := s.FetchEntry(uuid.New().String(), uuid.New().String())

	assert.NotNil(t, err, "expecting error")
	assert.Nil(t, res, "expecting nil result")
}
