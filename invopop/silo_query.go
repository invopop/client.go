package invopop

import (
	"context"
	"net/url"
	"path"
	"strconv"
)

const (
	queryPath = "query"
	countPath = "count"
)

// Queries against silo's read model differ from the plain entry list in two
// ways worth keeping in mind:
//
//   - Results are eventually consistent. An entry created a moment ago may not
//     appear until the collection's Watermark passes its creation time.
//   - Filters are expressed over a fixed vocabulary of fields per GOBL schema,
//     published in the API's OpenAPI specification.

// QuerySiloEntries is used to search entries of a single document type using
// a filter expression.
type QuerySiloEntries struct {
	Schema  string `query:"schema" title:"Schema" description:"GOBL document type to search." example:"bill/invoice"`
	Filter  string `query:"filter" title:"Filter" description:"Filter expression, combining terms with AND, and OR only between values of a single field." example:"state = \"sent\" AND issue_date > \"2026-01-01\""`
	OrderBy string `query:"order_by" title:"Order By" description:"Sort order, as a field name optionally followed by asc or desc." example:"issue_date desc"`
	Limit   int32  `query:"limit" title:"Limit" description:"Maximum number of entries in a page of results, up to 100." example:"20"`
	Cursor  string `query:"cursor" title:"Cursor" description:"Position provided by the previous result's next_cursor property. Only valid for the same schema, filter and order."`
	Fields  string `query:"fields" title:"Fields" description:"Comma-separated list of fields to populate on each entry." example:"id,code,state,issue_date"`
	Count   bool   `query:"count" title:"Count" description:"When true, the response includes the number of matching entries."`
}

// SiloQueryEntryCollection is a page of entries matching a structured query.
type SiloQueryEntryCollection struct {
	List []*SiloEntry `json:"list"`
	// Query
	Schema  string `json:"schema"`
	Filter  string `json:"filter,omitempty"`
	OrderBy string `json:"order_by,omitempty"`
	// Position
	Limit      int32  `json:"limit"`
	Cursor     string `json:"cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	// Count, only when requested
	Count            int32 `json:"count,omitempty"`
	CountCapped      bool  `json:"count_capped,omitempty"`
	CountUnavailable bool  `json:"count_unavailable,omitempty"`
	// Freshness and advice
	Watermark string   `json:"watermark,omitempty"`
	Notes     []string `json:"notes,omitempty"`
}

// CountSiloEntries is used to count the entries matching a query without
// returning them.
type CountSiloEntries struct {
	Schema string `query:"schema" title:"Schema" description:"GOBL document type to count." example:"bill/invoice"`
	Filter string `query:"filter" title:"Filter" description:"Filter expression, as used when querying entries." example:"state = \"sent\""`
}

// SiloEntryCount is the number of entries matching a query.
type SiloEntryCount struct {
	Schema           string `json:"schema"`
	Filter           string `json:"filter,omitempty"`
	Count            int32  `json:"count"`
	CountCapped      bool   `json:"count_capped,omitempty"`
	CountUnavailable bool   `json:"count_unavailable,omitempty"`
}

// Query searches silo entries of one document type using a filter expression.
// Pagination is supported using the collection's Cursor and NextCursor
// parameters.
func (svc *SiloEntriesService) Query(ctx context.Context, req *QuerySiloEntries) (*SiloQueryEntryCollection, error) {
	query := make(url.Values)
	query.Add("schema", req.Schema)
	if req.Filter != "" {
		query.Add("filter", req.Filter)
	}
	if req.OrderBy != "" {
		query.Add("order_by", req.OrderBy)
	}
	if req.Limit != 0 {
		query.Add("limit", strconv.Itoa(int(req.Limit)))
	}
	if req.Cursor != "" {
		query.Add("cursor", req.Cursor)
	}
	if req.Fields != "" {
		query.Add("fields", req.Fields)
	}
	if req.Count {
		query.Add("count", "true")
	}
	p := path.Join(siloBasePath, entriesPath, queryPath) + "?" + query.Encode()
	col := new(SiloQueryEntryCollection)
	return col, svc.client.get(ctx, p, col)
}

// Count provides the number of silo entries matching the query.
func (svc *SiloEntriesService) Count(ctx context.Context, req *CountSiloEntries) (*SiloEntryCount, error) {
	query := make(url.Values)
	query.Add("schema", req.Schema)
	if req.Filter != "" {
		query.Add("filter", req.Filter)
	}
	p := path.Join(siloBasePath, entriesPath, queryPath, countPath) + "?" + query.Encode()
	out := new(SiloEntryCount)
	return out, svc.client.get(ctx, p, out)
}
