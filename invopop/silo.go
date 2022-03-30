package invopop

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/org"
)

const (
	siloBasePath  = "/silo/v1"
	envelopesPath = "envelopes"
)

// SiloService implements the Invopop Silo API.
type SiloService service

// Envelope defines the fields provided by the Silo envelope end points.
type Envelope struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	Type   string       `json:"typ,omitempty"`
	Digest *dsig.Digest `json:"digest,omitempty"`
	Tags   []string     `json:"tags,omitempty"`
	Meta   org.Meta     `json:"meta,omitempty"`

	Attachments []*Attachment   `json:"attachments,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"` // may not always be available
}

// EnvelopeCollection contains a list of Envelopes that start from the provided created_at
// timestamp.
type EnvelopeCollection struct {
	List          []*Envelope `json:"list"`
	Limit         int32       `json:"limit"`
	CreatedAt     string      `json:"created_at"`
	NextCreatedAt string      `json:"next_created_at,omitempty"`
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

// ListEnvelopes provides a list of the envelopes that belong to the user. Pagination is supported
// using the EnvelopeCollections CreatedAt and NextCreatedAt parameters.
func (svc *SiloService) ListEnvelopes(ctx context.Context, col *EnvelopeCollection) error {
	p := path.Join(siloBasePath, envelopesPath)
	query := make(url.Values)
	if col.Limit != 0 {
		query.Add("limit", strconv.Itoa(int(col.Limit)))
	}
	if col.CreatedAt != "" {
		query.Add("created_at", col.CreatedAt)
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	return svc.client.get(ctx, p, col)
}

// CreateEnvelope sends the provided Envelope objects `data` to the server for storage.
// Only the `data` field and `ID` will be used.
func (svc *SiloService) CreateEnvelope(ctx context.Context, e *Envelope) error {
	return svc.client.put(ctx, path.Join(siloBasePath, envelopesPath, e.ID), e)
}

// FetchEnvelope updates the provided envelope instance with the results from the server.
func (svc *SiloService) FetchEnvelope(ctx context.Context, e *Envelope) error {
	return svc.client.get(ctx, path.Join(siloBasePath, envelopesPath, e.ID), e)
}
