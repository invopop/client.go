package invopop

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/dsig"
)

const (
	entriesPath = "entries"
)

// SiloEntriesService is responsible for managing the connection
// to the Silo API endpoints.
type SiloEntriesService service

// SiloEntry defines the fields provided by the Silo entry end points.
type SiloEntry struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Folder    string       `json:"folder" title:"Folder" description:"Key for the folder where the entry is located." example:"sales"`
	State     string       `json:"state,omitempty" title:"State" description:"Current state of the silo entry if not a draft." example:"sent"`
	Draft     bool         `json:"draft,omitempty" title:"Draft" description:"When true, indicates that the envelope is a draft." example:"true"`
	Invalid   bool         `json:"invalid,omitempty" title:"Invalid" description:"When true, the envelope's contents are invalid and need to be reviewed." example:"true"`
	EnvSchema string       `json:"env_schema" title:"Envelope Schema" description:"Schema URL for the envelope." example:"https://gobl.org/draft-0/envelope"`
	DocSchema string       `json:"doc_schema" title:"Object Schema" description:"Schema URL for the envelope's payload." example:"https://gobl.org/draft-0/bill/invoice"`
	Digest    *dsig.Digest `json:"digest" title:"Digest" description:"A copy of the digest from the envelope."`
	Tags      []string     `json:"tags,omitempty" title:"Tags" x-order:"8"`
	Context   string       `json:"context,omitempty" title:"Context" description:"When entry provided within a related query, this is the context within the document." example:"line.item"`

	Snippet json.RawMessage `json:"snippet,omitempty"`

	Attachments []*SiloAttachment `json:"attachments,omitempty"`
	Data        json.RawMessage   `json:"data,omitempty"` // may not always be available
	Meta        []*SiloMeta       `json:"meta,omitempty" title:"Meta" description:"Additional meta fields associated with the entry."`
}

// SiloEntryCollection contains a list of Entries that start from the provided created_at
// timestamp.
type SiloEntryCollection struct {
	List []*SiloEntry `json:"list"`
	// Filters
	Folder    string `json:"folder,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	// Position
	Limit      int32  `json:"limit"`
	Cursor     string `json:"cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// SiloAttachment represents a file that was generated by one of the job's intents
// and is stored by the Silo service.
type SiloAttachment struct {
	ID        string `json:"id" title:"ID" description:"The UUID that identifies the attachment."`
	CreatedAt string `json:"created_at" title:"Created At" description:"The date and time this attachment was created." example:"2018-01-01T00:00:00.000Z"`

	Name   string `json:"name" title:"Name" description:"Name of the file." example:"invoice.pdf"`
	Desc   string `json:"desc,omitempty" title:"Description" description:"Description of the file." example:"Invoice for January 2021."`
	Hash   string `json:"hash" title:"Hash" description:"SHA256 hash of the file." example:"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"`
	MIME   string `json:"mime" title:"MIME" description:"MIME type of the file." example:"application/pdf"`
	Size   int32  `json:"size" title:"Size" description:"Size of the file in bytes." example:"12345"`
	Stored bool   `json:"stored" title:"Stored" description:"When true, indicates that the file's contents have been uploaded successfully." example:"true"`
	URL    string `json:"url" title:"Public URL" description:"URL where the file's contents can be downloaded."` // public URL

	Meta map[string]string `json:"meta,omitempty" title:"Meta" description:"Any additional meta data about the attachment."`
}

// CreateSiloEntry defines the required fields to create an envelope in the Silo.
type CreateSiloEntry struct {
	ID           string          `json:"-"`
	Folder       string          `json:"folder,omitempty" title:"Folder" description:"In which folder the entry should be associated, leave empty to use the automatic rules."`
	Data         json.RawMessage `json:"data,omitempty" title:"Data" description:"Data contents to upload which may either be a GOBL Envelope or Object. Any partial data will be calculated and validated automatically."`
	PreviousID   string          `json:"previous_id,omitempty" title:"Previous ID" description:"The UUIDv1 of the previous silo entry to copy." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	ContentType  string          `json:"content_type,omitempty" title:"Content Type" description:"The content type of the data being uploaded." example:"application/json"`
	Correct      json.RawMessage `json:"correct,omitempty" title:"Correct" description:"JSON object containing the GOBL correction option data." example:"{\"credit\": true}"`
	AllowInvalid bool            `json:"allow_invalid,omitempty" title:"Allow Invalid" description:"When true, the envelope's contents are allowed to be invalid. Only applied for enrolled applications." example:"true"`
}

// UpdateSiloEntry allows for a silo document to be updated under certain conditions.
type UpdateSiloEntry struct {
	ID     string          `json:"-"`
	Folder string          `json:"folder,omitempty" title:"Folder" description:"New location for the silo entry." example:"drafts"`
	Data   json.RawMessage `json:"data" title:"Data" description:"Updated envelope data. New document contents will only be accepted if the existing envelope data is still a draft."`
}

// FindSiloEntries is used to list entries ordered by date.
type FindSiloEntries struct {
	Folder    string `query:"folder" title:"Folder" description:"Folder to search within." example:"invoices"`
	CreatedAt string `query:"created_at" title:"Created At" description:"Date from which results are provided." example:"2023-08-02T00:00:00.000Z"`
	Cursor    string `query:"cursor" title:"Cursor" description:"Position provided by the previous result's next_cursor property."`
	Limit     int32  `query:"limit" title:"Limit" description:"Maximum number of entries to show in a page of results." example:"20"`
}

// List provides a list of the silo entries that belong to the user. Pagination is supported
// using the EntryCollection's Cursor and NextCursor parameters.
func (svc *SiloEntriesService) List(ctx context.Context, req *FindSiloEntries) (*SiloEntryCollection, error) {
	p := path.Join(siloBasePath, entriesPath)
	query := make(url.Values)
	if req.Limit != 0 {
		query.Add("limit", strconv.Itoa(int(req.Limit)))
	}
	if req.CreatedAt != "" {
		query.Add("created_at", req.CreatedAt)
	}
	if req.Cursor != "" {
		query.Add("cursor", req.Cursor)
	}
	if req.Folder != "" {
		query.Add("folder", req.Folder)
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	col := new(SiloEntryCollection)
	return col, svc.client.get(ctx, p, col)
}

// Fetch loads the requested silo entry by its ID.
func (svc *SiloEntriesService) Fetch(ctx context.Context, id string) (*SiloEntry, error) {
	e := new(SiloEntry)
	return e, svc.client.get(ctx, path.Join(siloBasePath, entriesPath, id), e)
}

// Create makes a request to persist a new silo entry.
func (svc *SiloEntriesService) Create(ctx context.Context, req *CreateSiloEntry) (*SiloEntry, error) {
	e := new(SiloEntry)
	return e, svc.client.put(ctx, path.Join(siloBasePath, entriesPath, req.ID), req, e)
}

// Update sends the provided Entry object `data` to the server for storage,
// updating the existing envelope.
func (svc *SiloEntriesService) Update(ctx context.Context, req *UpdateSiloEntry) (*SiloEntry, error) {
	e := new(SiloEntry)
	return e, svc.client.patch(ctx, path.Join(siloBasePath, entriesPath, req.ID), req, e)
}

// Envelope provides the silo entry's data as a GOBL envelope.
func (se *SiloEntry) Envelope() (*gobl.Envelope, error) {
	env := new(gobl.Envelope)
	if err := json.Unmarshal(se.Data, env); err != nil {
		return nil, err
	}
	return env, nil
}
