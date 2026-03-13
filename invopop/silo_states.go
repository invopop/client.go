package invopop

import (
	"context"
	"path"
)

const (
	statesPath = "states"
)

// SiloStatesService provides access to the Silo States API endpoints.
type SiloStatesService service

// SiloState represents a state transition for a silo entry.
type SiloState struct {
	ID        string            `json:"id,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	Key       string            `json:"key" title:"Key" description:"Key identifying the state." example:"sent"`
	Src       string            `json:"src,omitempty" title:"Src" description:"Identifier for the source of the state."`
	SrcID     string            `json:"src_id,omitempty" title:"Src ID" description:"UUID for the entity that created the state."`
	Notes     string            `json:"notes,omitempty" title:"Notes" description:"Additional notes about the state transition."`
	Faults    []*Fault          `json:"faults,omitempty" title:"Faults" description:"List of faults that occurred during processing."`
	Meta      map[string]string `json:"meta,omitempty" title:"Meta" description:"Additional structured fields."`
}

// SiloStateCollection contains a list of states for a silo entry.
type SiloStateCollection struct {
	List []*SiloState `json:"list"`
}

// CreateSiloState defines the fields required to create a new state for a silo entry.
type CreateSiloState struct {
	EntryID string            `json:"-"`
	At      string            `json:"-"` // ISO timestamp for idempotent PUT, leave empty for POST
	Key     string            `json:"key" title:"Key" description:"Key identifying the state." example:"sent"`
	Notes   string            `json:"notes,omitempty" title:"Notes" description:"Additional notes about the state transition."`
	Meta    map[string]string `json:"meta,omitempty" title:"Meta" description:"Additional structured fields."`
}

// List provides a list of states for the given silo entry.
func (svc *SiloStatesService) List(ctx context.Context, entryID string) (*SiloStateCollection, error) {
	p := path.Join(siloBasePath, entriesPath, entryID, statesPath)
	col := new(SiloStateCollection)
	return col, svc.client.get(ctx, p, col)
}

// Create sends a request to create a new state for the given silo entry.
// If At is provided, uses an idempotent PUT with the timestamp; otherwise uses POST.
func (svc *SiloStatesService) Create(ctx context.Context, req *CreateSiloState) (*SiloState, error) {
	s := new(SiloState)
	if req.At != "" {
		p := path.Join(siloBasePath, entriesPath, req.EntryID, statesPath, req.At)
		return s, svc.client.put(ctx, p, req, s)
	}
	p := path.Join(siloBasePath, entriesPath, req.EntryID, statesPath)
	return s, svc.client.post(ctx, p, req, s)
}
