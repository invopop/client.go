package snippets_test

import (
	"encoding/json"
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

func TestParseSchemas(t *testing.T) {
	t.Run("with note message", func(t *testing.T) {
		snippet := snippets.Parse(
			"https://gobl.org/draft-0/note/message",
			[]byte(`{"uuid":"0192cd23-5c2a-7000-8000-000000000001","title":"Hello"}`),
		)
		require.NotNil(t, snippet)

		msg, ok := snippet.(*snippets.NoteMessage)
		require.True(t, ok)
		assert.Equal(t, "Hello", msg.Title)
		assert.Equal(t, "0192cd23-5c2a-7000-8000-000000000001", msg.UUID)
	})

	t.Run("with bill invoice", func(t *testing.T) {
		data := []byte(`{
			"type": "standard",
			"series": "F1",
			"code": "0123",
			"currency": "EUR",
			"issue_date": "2024-11-01",
			"total": "100.00",
			"tax": "21.00",
			"total_with_tax": "121.00",
			"payable": "121.00",
			"supplier": {"name": "Supplier SL", "country": "ES", "tax_code": "B98602642"},
			"customer": {"name": "Customer Ltd", "country": "GB"}
		}`)
		snippet := snippets.Parse("https://gobl.org/draft-0/bill/invoice", data)
		require.NotNil(t, snippet)

		inv, ok := snippet.(*snippets.BillInvoice)
		require.True(t, ok)
		assert.Equal(t, "standard", inv.Type)
		assert.Equal(t, "F1", inv.Series)
		assert.Equal(t, "0123", inv.Code)
		assert.Equal(t, "EUR", inv.Currency.String())
		assert.Equal(t, "2024-11-01", inv.IssueDate.String())

		require.NotNil(t, inv.Total)
		assert.Equal(t, "100.00", inv.Total.String())
		require.NotNil(t, inv.Tax)
		assert.Equal(t, "21.00", inv.Tax.String())
		require.NotNil(t, inv.TotalWithTax)
		assert.Equal(t, "121.00", inv.TotalWithTax.String())
		require.NotNil(t, inv.Payable)
		assert.Equal(t, "121.00", inv.Payable.String())

		require.NotNil(t, inv.Supplier)
		assert.Equal(t, "Supplier SL", inv.Supplier.Name)
		assert.Equal(t, "ES", inv.Supplier.Country.String())
		assert.Equal(t, "B98602642", inv.Supplier.TaxCode)

		require.NotNil(t, inv.Customer)
		assert.Equal(t, "GB", inv.Customer.Country.String())
	})

	t.Run("with an unknown schema", func(t *testing.T) {
		assert.Nil(t, snippets.Parse("https://gobl.org/draft-0/pay/advance", []byte(`{}`)))
		assert.Nil(t, snippets.Parse("", []byte(`{}`)))
	})

	t.Run("with invalid JSON data", func(t *testing.T) {
		assert.Nil(t, snippets.Parse("https://gobl.org/draft-0/org/party", []byte(`not json`)))
		assert.Nil(t, snippets.Parse("https://gobl.org/draft-0/bill/invoice", nil))
	})

	t.Run("matches on the schema suffix only", func(t *testing.T) {
		// Versioned or relocated schema hosts should still resolve.
		snippet := snippets.Parse("https://example.com/other/org/party", []byte(`{"name":"Jane"}`))
		require.NotNil(t, snippet)
		assert.IsType(t, &snippets.OrgParty{}, snippet)
	})
}

func TestSnippetRoundTrip(t *testing.T) {
	// The snippet types are returned directly by the API, so their JSON
	// representation forms part of the client's contract.
	data := []byte(`{"type":"standard","code":"0123","currency":"EUR","issue_date":"2024-11-01","total_with_tax":"121.00"}`)
	snippet := snippets.Parse("https://gobl.org/draft-0/bill/invoice", data)
	require.NotNil(t, snippet)

	out, err := json.Marshal(snippet)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "standard",
		"code": "0123",
		"currency": "EUR",
		"issue_date": "2024-11-01",
		"total_with_tax": "121.00"
	}`, string(out))
}
