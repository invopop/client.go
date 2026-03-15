package invopop

import (
	"context"
	"net/url"
	"path"
	"strconv"
)

const (
	searchPath = "search"
)

// SearchSiloEntries defines the parameters for searching silo entries.
type SearchSiloEntries struct {
	Query  string `query:"q" title:"Query" description:"Search query string." example:"invoice"`
	Folder string `query:"folder" title:"Folder" description:"Optional folder to filter results." example:"sales"`
	Limit  int32  `query:"limit" title:"Limit" description:"Maximum number of results to return." example:"20"`
	Offset int32  `query:"offset" title:"Offset" description:"Pagination offset for results." example:"0"`
}

// SiloEntrySearchCollection contains a list of silo entries matching a search query.
type SiloEntrySearchCollection struct {
	List   []*SiloEntry `json:"list"`
	Folder string       `json:"folder"`
	Query  string       `json:"query"`
	Limit  int32        `json:"limit"`
	Offset int32        `json:"offset"`
}

// Search performs a search across silo entries.
func (svc *SiloService) Search(ctx context.Context, req *SearchSiloEntries) (*SiloEntrySearchCollection, error) {
	p := path.Join(siloBasePath, searchPath)
	query := make(url.Values)
	if req.Query != "" {
		query.Add("q", req.Query)
	}
	if req.Folder != "" {
		query.Add("folder", req.Folder)
	}
	if req.Limit != 0 {
		query.Add("limit", strconv.Itoa(int(req.Limit)))
	}
	if req.Offset != 0 {
		query.Add("offset", strconv.Itoa(int(req.Offset)))
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	col := new(SiloEntrySearchCollection)
	return col, svc.client.get(ctx, p, col)
}
