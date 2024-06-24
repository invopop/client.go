// Package snippets provides helpers to generate GOBL document snippets.
package snippets

import (
	"encoding/json"
	"strings"
)

// Parse reads in the data and returns a snippet object that is appropriate for
// the schema. This is assumed to be used in non-essential scenarios, so there
// is no error handling.
func Parse(schema string, data []byte) any {
	var obj any
	switch {
	case strings.HasSuffix(schema, "org/party"):
		obj = new(OrgParty)
	case strings.HasSuffix(schema, "note/message"):
		obj = new(NoteMessage)
	case strings.HasSuffix(schema, "bill/invoice"):
		obj = new(BillInvoice)
	default:
		return nil
	}
	if err := json.Unmarshal(data, obj); err != nil {
		return nil
	}
	return obj
}
