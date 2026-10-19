package invopop

import (
	"context"
	"errors"
	"io"
	"path"
	"strconv"

	"resty.dev/v3"
)

const (
	entriesExportPath = "export"
	entriesImportPath = "import"
)

// Categories of an entry's files that an export may embed.
const (
	ExportAttachmentsAll        = "all"
	ExportAttachmentsFormat     = FileCategoryFormat
	ExportAttachmentsAttachment = FileCategoryAttachment
)

// ExportSiloEntry converts a silo entry's GOBL envelope into an external
// format. Nothing is stored in the entry.
type ExportSiloEntry struct {
	// ID of the silo entry to export.
	ID string
	// Formats to export to in order of preference: the first that accepts the
	// document is used.
	Formats []string
	// Attachments is the category of the entry's files to embed, one of the
	// ExportAttachments constants. None are embedded when empty.
	Attachments string
	// MIME limits the files embedded to these types, any when empty.
	MIME []string
}

// ImportSiloEntry creates a new silo entry from a document in an external
// format, such as UBL or CII. The original document is kept as one of the
// entry's files, along with any files embedded in it.
type ImportSiloEntry struct {
	// ID of the silo entry to create, assigned automatically when empty.
	ID string
	// Key used to identify the entry idempotently within a workspace:
	// repeating an import with the same key provides the entry already
	// created.
	Key string
	// Folder for the entry, leave empty to use the automatic rules.
	Folder string
	// Formats to limit detection to, any when empty.
	Formats []string
	// Name of the original document's file, e.g. "invoice.xml".
	Name string
	// AllowInvalid allows the envelope's contents to be invalid.
	AllowInvalid bool
	// Sign the GOBL envelope automatically.
	Sign bool
	// Private stores the entry's files as private, and does not link
	// attachments from the document.
	Private bool
	// MIME type of the document, sent as the request's content type.
	MIME string
	// Data provides the raw document.
	Data io.Reader
}

// Export converts a silo entry into the first of the requested formats that
// accepts it. The result is streamed: close it once read.
func (svc *SiloEntriesService) Export(ctx context.Context, req *ExportSiloEntry) (*SiloExport, error) {
	if req.ID == "" {
		return nil, errors.New("missing id")
	}
	if len(req.Formats) == 0 {
		return nil, errors.New("missing formats")
	}
	q := formatsQuery(req.Formats)
	if req.Attachments != "" {
		q.Add("attachments", req.Attachments)
	}
	for _, m := range req.MIME {
		q.Add("mime", m)
	}
	p := path.Join(siloBasePath, entriesPath, req.ID, entriesExportPath) + "?" + q.Encode()
	return sendExport(svc.client.conn.R().SetContext(ctx), resty.MethodGet, p)
}

// Import creates a new silo entry from a document in an external format. The
// document is streamed to the server as it is read. An existing entry is never
// updated.
func (svc *SiloEntriesService) Import(ctx context.Context, req *ImportSiloEntry) (*SiloEntry, error) {
	if req.Data == nil {
		return nil, errors.New("missing data")
	}
	q := formatsQuery(req.Formats)
	if req.Key != "" {
		q.Add("key", req.Key)
	}
	if req.Folder != "" {
		q.Add("folder", req.Folder)
	}
	if req.Name != "" {
		q.Add("name", req.Name)
	}
	if req.AllowInvalid {
		q.Add("allow_invalid", strconv.FormatBool(true))
	}
	if req.Sign {
		q.Add("sign", strconv.FormatBool(true))
	}
	if req.Private {
		q.Add("private", strconv.FormatBool(true))
	}

	e := new(SiloEntry)
	if req.ID == "" {
		p := path.Join(siloBasePath, entriesPath, entriesImportPath)
		if len(q) > 0 {
			p = p + "?" + q.Encode()
		}
		return e, svc.client.postRaw(ctx, p, rawMIME(req.MIME), req.Data, e)
	}
	p := path.Join(siloBasePath, entriesPath, req.ID, entriesImportPath)
	if len(q) > 0 {
		p = p + "?" + q.Encode()
	}
	return e, svc.client.putRaw(ctx, p, rawMIME(req.MIME), req.Data, e)
}

// rawMIME defaults the content type of a raw request body.
func rawMIME(mime string) string {
	if mime == "" {
		return defaultFileMIME
	}
	return mime
}
