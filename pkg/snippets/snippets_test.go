package snippets_test

import (
	"testing"

	"github.com/invopop/client.go/pkg/snippets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("with org party", func(t *testing.T) {
		schema := "https://gobl.org/draft-0/org/party"
		data := []byte(`{"name":"John Doe"}`)

		snippet := snippets.Parse(schema, data)
		require.NotNil(t, snippet)
		assert.IsType(t, &snippets.OrgParty{}, snippet)

		orgParty, ok := snippet.(*snippets.OrgParty)
		assert.True(t, ok)
		require.NotNil(t, orgParty)
		assert.Equal(t, "John Doe", orgParty.Name)
	})
}
