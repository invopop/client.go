package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"path"
)

const (
	metaPath = "meta"
)

// SiloMetaService provides access to meta endpoints. Access to these requires
// the special "enrolled" scope in tokens.
type SiloMetaService service

// SiloMeta describes a meta row embedded inside a Silo Entry.
type SiloMeta struct {
	ID        string `json:"id" title:"ID" description:"Compound ID of the meta row." example:"347c5b04-cde2-11ed-afa1-0242ac120002:source:key"`
	CreatedAt string `json:"created_at" title:"Created At" description:"Timestamp of when the row was created." example:"2018-01-01T00:00:00.000Z"`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"Timestamp of when the row was last updated." example:"2018-01-01T00:00:00.000Z"`

	OwnerID   string          `json:"owner_id,omitempty" title:"Owner ID" description:"UUID of the owner of the silo entry, typically a workspace ID. Included for reference when the authentication token does not already include the owner such as for lookups by reference."`
	EntryID   string          `json:"entry_id,omitempty" title:"Entry ID" description:"ID of the entry this meta row belongs to" example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	Src       string          `json:"src,omitempty" title:"Source" description:"The service or source that create this meta entry." example:"provider"`
	Key       string          `json:"key,omitempty" title:"Key" description:"Key used to identify the meta entry by the source." example:"service-id"`
	Ref       string          `json:"ref,omitempty" title:"Reference" description:"Indexable value used to locate the meta entry if a silo entry ID is not available."`
	Value     json.RawMessage `json:"value,omitempty" title:"Value" description:"The JSON data stored with the meta row." example:"{\"key\": \"value\"}"`
	LinkURL   string          `json:"link_url,omitempty" title:"Link URL" description:"URL associated with the meta row that may be used to perform additional actions or view more details." example:"https://example.com/info"`
	LinkScope string          `json:"link_scope,omitempty" title:"Link Scope" description:"Describes the context in which the link should be made available." example:"public"`
	Shared    bool            `json:"shared,omitempty" title:"Shared" description:"When true, the meta entry can be shared with other applications." example:"true"`
}

// UpsertSiloMeta is to update or create a new meta row for an entry.
type UpsertSiloMeta struct {
	EntryID   string          `json:"-"`
	Key       string          `json:"-"`
	Ref       string          `json:"ref,omitempty" title:"Reference" description:"Indexable value used to locate the meta entry if a silo entry ID is not available."`
	Value     json.RawMessage `json:"value,omitempty" title:"Value" description:"The JSON data stored with the meta row." example:"{\"key\": \"value\"}"`
	LinkURL   string          `json:"url,omitempty" title:"Link URL" description:"URL associated with the meta row that may be used to perform additional actions or view more details." example:"https://example.com/info"`
	LinkScope string          `json:"scope,omitempty" title:"Link Scope" description:"Describes the context in which the link should be made available." example:"public"`
	Indexed   bool            `json:"indexed,omitempty" title:"Indexed" description:"When true, the meta entry is indexed for search." example:"true"`
	Secure    bool            `json:"secure,omitempty" title:"Secure" description:"When true, the meta entry is never included in lists and needs to be specifically requested." example:"true"`
	Shared    bool            `json:"shared,omitempty" title:"Shared" description:"When true, the meta entry can be shared with other applications." example:"true"`
}

// Fetch retrieves a meta row by its key.
func (s *SiloMetaService) Fetch(ctx context.Context, entryID, key string) (*SiloMeta, error) {
	if entryID == "" {
		return nil, errors.New("missing entry ID")
	}
	if key == "" {
		return nil, errors.New("missing key")
	}
	m := new(SiloMeta)
	return m, s.client.get(ctx, path.Join(siloBasePath, entriesPath, entryID, metaPath, key), m)
}

// FetchByRef retrieves a meta row by its reference value. The ref is used as an
// alternative to the entry ID.
func (s *SiloMetaService) FetchByRef(ctx context.Context, key, ref string) (*SiloMeta, error) {
	if key == "" {
		return nil, errors.New("missing key")
	}
	if ref == "" {
		return nil, errors.New("missing ref")
	}
	p := path.Join(siloBasePath, entriesPath, metaPath, key, ref)
	m := new(SiloMeta)
	return m, s.client.get(ctx, p, m)
}

// Upsert will either create a new meta row or update an existing one. The key of the SiloMeta
// row will be used to upload.
func (s *SiloMetaService) Upsert(ctx context.Context, req *UpsertSiloMeta) (*SiloMeta, error) {
	if req.Key == "" {
		return nil, errors.New("missing key")
	}
	p := path.Join(siloBasePath, entriesPath, req.EntryID, metaPath, req.Key)
	m := new(SiloMeta)
	return m, s.client.put(ctx, p, req, m)
}

// Delete will delete a meta row by its key.
func (s *SiloMetaService) Delete(ctx context.Context, entryID, key string) (*SiloMeta, error) {
	if entryID == "" {
		return nil, errors.New("missing entry ID")
	}
	if key == "" {
		return nil, errors.New("missing key")
	}
	m := new(SiloMeta)
	return m, s.client.delete(ctx, path.Join(siloBasePath, entriesPath, entryID, metaPath, key), m)
}
