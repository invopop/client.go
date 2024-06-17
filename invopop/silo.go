package invopop

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/dsig"
)

const (
	siloBasePath = "/silo/v1"
	entriesPath  = "entries"
)

// SiloService implements the Invopop Silo API.
type SiloService service

// Entry defines the fields provided by the Silo entry end points.
type Entry struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Folder    string       `json:"folder,omitempty"`
	EnvSchema string       `json:"env_schema,omitempty"`
	DocSchema string       `json:"doc_schema,omitempty"`
	Digest    *dsig.Digest `json:"digest,omitempty"`
	Tags      []string     `json:"tags,omitempty"`
	Meta      cbc.Meta     `json:"meta,omitempty"`
	Draft     bool         `json:"draft,omitempty"`

	Snippet json.RawMessage `json:"snippet,omitempty"`

	Attachments []*Attachment   `json:"attachments,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"` // may not always be available
}

// EntryCollection contains a list of Entries that start from the provided created_at
// timestamp.
type EntryCollection struct {
	List []*Entry `json:"list"`
	// Filters
	Folder    string `json:"folder,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	// Position
	Limit      int32  `json:"limit"`
	Cursor     string `json:"cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// Attachment represents a file that was generated by one of the job's intents
// and is stored by the Silo service.
type Attachment struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Desc string            `json:"desc,omitempty"`
	Hash string            `json:"hash"`
	MIME string            `json:"mime"`
	Size int32             `json:"size"`
	URL  string            `json:"url"` // public URL
	Meta map[string]string `json:"meta,omitempty"`
}

// ListEntries provides a list of the entries that belong to the user. Pagination is supported
// using the EntryCollection's Cursor and NextCursor parameters.
func (svc *SiloService) ListEntries(ctx context.Context, col *EntryCollection) error {
	p := path.Join(siloBasePath, entriesPath)
	query := make(url.Values)
	if col.Limit != 0 {
		query.Add("limit", strconv.Itoa(int(col.Limit)))
	}
	if col.CreatedAt != "" {
		query.Add("created_at", col.CreatedAt)
	}
	if col.Cursor != "" {
		query.Add("cursor", col.Cursor)
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	return svc.client.get(ctx, p, col)
}

// CreateEntry sends the provided Entry objects `data` to the server for storage.
// Only the `data` field and `ID` will be used.
func (svc *SiloService) CreateEntry(ctx context.Context, e *Entry) error {
	return svc.client.put(ctx, path.Join(siloBasePath, entriesPath, e.ID), e, e)
}

// UpdateEntry sends the provided Entry object `data` to the server for storage,
// updating the existing envelope.
func (svc *SiloService) UpdateEntry(ctx context.Context, e *Entry) error {
	return svc.client.patch(ctx, path.Join(siloBasePath, entriesPath, e.ID), e, e)
}

// FetchEnvelope updates the provided envelope instance with the results from the server.
func (svc *SiloService) FetchEnvelope(ctx context.Context, e *Entry) error {
	return svc.client.get(ctx, path.Join(siloBasePath, entriesPath, e.ID), e)
}
