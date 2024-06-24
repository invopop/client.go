package invopop

import (
	"context"
	"path"
)

const (
	sequenceBasePath  = "/sequence/v1"
	seriesPath        = "series"
	seriesEntriesPath = "entries"
)

// SequenceService handles communication with the Invopop
// sequences API end points.
type SequenceService service

// Series defines a instance of a code series managed by the
// sequence service.
type Series struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Prefix  string `json:"prefix,omitempty"`
	Padding int32  `json:"padding,omitempty"`
	Suffix  string `json:"suffix,omitempty"`
	Start   int32  `json:"start,omitempty"`

	LastIndex   int64  `json:"last_index,omitempty"`
	LastEntryID string `json:"last_entry_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// SeriesCollection contains a list of series.
type SeriesCollection struct {
	List []*Series `json:"list"`
}

// SeriesEntry represents a single instance of an entry inside
// a series.
type SeriesEntry struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

// CreateSeries defines the expected fields to create a new series entry.
type CreateSeries struct {
	ID          string `json:"-"`
	Name        string `json:"name" title:"Name" description:"The name of the series." example:"My Series"`
	Description string `json:"description" title:"Description" description:"A description of the series." example:"This series is used for sales."`
	Code        string `json:"code" title:"Code" description:"A code that can be used to identify the series." example:"SALES-2020"`
	Prefix      string `json:"prefix" title:"Prefix" description:"A prefix that will be prepended to all entries." example:"INV-"`
	Padding     int32  `json:"padding" title:"Padding" description:"The number of 0s to pad the generated codes with." example:"5"`
	Suffix      string `json:"suffix" title:"Suffix" description:"A suffix that will be appended to all entries." example:"-F1"`
	Start       int32  `json:"start" title:"Start" description:"The starting index for the series." example:"1"`
}

// CreateSeriesEntry is used to create a new entry in a series.
type CreateSeriesEntry struct {
	ID   string            `json:"id" path:"id" title:"ID" description:"The UUID (any version) of the entry to create." example:"a8904315-3d16-4a95-91c1-30d6cdde553e"`
	Meta map[string]string `json:"meta,omitempty" title:"Meta" description:"A set of key/value pairs that can be used to store additional information about the entry." example:"{ \"name\": \"John Doe\" }"`
	Sig  string            `json:"sig,omitempty" title:"Signature" description:"JSON Web Signature of the key properties used to create the entry."`
}

// List will populate the series collection with the series that match the
// conditions, if any.
func (svc *SequenceService) List(ctx context.Context) (*SeriesCollection, error) {
	p := path.Join(sequenceBasePath, seriesPath)
	m := new(SeriesCollection)
	return m, svc.client.get(ctx, p, m)
}

// Fetch requests a specific sequence by its ID.
func (svc *SequenceService) Fetch(ctx context.Context, id string) (*Series, error) {
	p := path.Join(sequenceBasePath, seriesPath, id)
	m := new(Series)
	return m, svc.client.get(ctx, p, m)
}

// Create will create the series.
func (svc *SequenceService) Create(ctx context.Context, req *CreateSeries) (*Series, error) {
	p := path.Join(sequenceBasePath, seriesPath, req.ID)
	m := new(Series)
	return m, svc.client.put(ctx, p, req, m)
}

// FetchEntry loads the series entry from the API by its ID.
func (svc *SequenceService) FetchEntry(ctx context.Context, seriesID, id string) (*SeriesEntry, error) {
	p := path.Join(sequenceBasePath, seriesPath, seriesID, seriesEntriesPath, id)
	m := new(SeriesEntry)
	return m, svc.client.get(ctx, p, m)
}

// CreateEntry will send a request to create a new series entry.
func (svc *SequenceService) CreateEntry(ctx context.Context, seriesID string, req *CreateSeriesEntry) (*SeriesEntry, error) {
	p := path.Join(sequenceBasePath, seriesPath, seriesID, seriesEntriesPath, req.ID)
	m := new(SeriesEntry)
	return m, svc.client.put(ctx, p, req, m)
}
