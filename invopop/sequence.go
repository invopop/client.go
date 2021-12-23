package invopop

import (
	"context"
	"fmt"
)

const (
	sequenceBasePath = "/sequence/v1"
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

// ListSeries will populate the series collection with the series that match the
// conditions, if any.
func (svc *SequenceService) ListSeries(ctx context.Context, sc *SeriesCollection) error {
	path := sequenceBasePath + "/series"
	return svc.client.get(ctx, path, sc)
}

// CreateSeries will create the series.
func (svc *SequenceService) CreateSeries(ctx context.Context, s *Series) error {
	path := fmt.Sprintf("%s/series/%s", sequenceBasePath, s.ID)
	return svc.client.put(ctx, path, s)
}

func (svc *SequenceService) CreateEntry(ctx context.Context, seriesID string, se *SeriesEntry) error {
	path := fmt.Sprintf("%s/series/%s/entries/%s", sequenceBasePath, seriesID, se.ID)
	return svc.client.put(ctx, path, se)
}