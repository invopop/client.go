package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"path"

	"github.com/invopop/gobl"
	"resty.dev/v3"
)

const (
	goblConvertPath        = "convert"
	goblConvertFormatsPath = "formats"
	goblConvertExportPath  = "export"
	goblConvertImportPath  = "import"

	// ConvertFormatHeader is the response header naming the format an export
	// used, as several may be offered.
	ConvertFormatHeader = "GOBL-Convert-Format"

	// ConvertDirectionImport filters formats that can be imported.
	ConvertDirectionImport = "import"
	// ConvertDirectionExport filters formats that can be exported to.
	ConvertDirectionExport = "export"
)

// GOBLFormat describes an external document format, such as UBL or CII, that
// GOBL documents can be converted to or from.
type GOBLFormat struct {
	// Key identifies the format in export and import requests.
	Key string `json:"key" title:"Key" example:"ubl+peppol+bis-v3"`
	// Name of the format, by language.
	Name map[string]string `json:"name,omitempty" title:"Name"`
	// Description of the format, by language.
	Description map[string]string `json:"description,omitempty" title:"Description"`
	// MIME type of documents in the format.
	MIME string `json:"mime,omitempty" title:"MIME" example:"application/xml"`
	// Syntax of the underlying document.
	Syntax string `json:"syntax,omitempty" title:"Syntax" example:"ubl"`
	// Countries the format applies in, empty for any.
	Countries []string `json:"countries,omitempty" title:"Countries"`
	// GOBL schemas documents in the format can be imported into.
	ImportSchemas []string `json:"import_schemas,omitempty" title:"Import Schemas"`
	// GOBL schemas that can be exported to the format.
	ExportSchemas []string `json:"export_schemas,omitempty" title:"Export Schemas"`
	// Attachments is true when exports can embed attachments.
	Attachments bool `json:"attachments,omitempty" title:"Attachments"`
}

// GOBLFormatCollection lists conversion formats.
type GOBLFormatCollection struct {
	List []*GOBLFormat `json:"list"`
}

// FetchGOBLFormats filters the conversion formats listed. All fields are
// optional.
type FetchGOBLFormats struct {
	// Country the formats apply in, e.g. "DE".
	Country string
	// Schema of the GOBL documents the formats convert, e.g.
	// "https://gobl.org/draft-0/bill/invoice".
	Schema string
	// Direction of the conversion, ConvertDirectionImport or
	// ConvertDirectionExport.
	Direction string
}

// ExportGOBL converts a GOBL document into an external format without storing
// anything.
type ExportGOBL struct {
	// Formats to export to in order of preference: the first that accepts the
	// document is used.
	Formats []string
	// Data provides the GOBL Envelope or Object JSON.
	Data io.Reader
	// Attachments to embed in the exported document.
	Attachments []*ExportGOBLAttachment
}

// ExportGOBLAttachment is a file to embed in an exported document.
type ExportGOBLAttachment struct {
	// ID identifies the attachment in the exported document, such as the
	// reference of a UBL or CII additional document.
	ID string
	// Name of the file, e.g. "receipt.pdf". Required.
	Name string
	// Desc describes what the file contains.
	Desc string
	// MIME type of the file, detected by the server when empty.
	MIME string
	// Data provides the file's contents.
	Data io.Reader
}

// ImportGOBL converts a document in an external format into a GOBL envelope
// without storing anything.
type ImportGOBL struct {
	// Formats to limit detection to, any when empty.
	Formats []string
	// MIME type of the document, sent as the request's content type.
	MIME string
	// Data provides the raw document.
	Data io.Reader
}

// GOBLImport is the result of importing a document.
type GOBLImport struct {
	// Format detected.
	Format string `json:"format"`
	// Data contains the GOBL envelope JSON built from the document.
	Data json.RawMessage `json:"data"`
	// Attachments embedded in the document.
	Attachments []*GOBLImportAttachment `json:"attachments,omitempty"`
}

// GOBLImportAttachment is a file that was embedded in an imported document.
type GOBLImportAttachment struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Desc string `json:"desc,omitempty"`
	MIME string `json:"mime,omitempty"`
	Size int64  `json:"size"`
	Data []byte `json:"data"`
}

// SiloExport is a document exported to an external format, along with the
// format used. Close it once read.
type SiloExport struct {
	Download
	// Format used for the export.
	Format string
}

// Formats lists the external formats GOBL documents can be converted to or
// from.
func (svc *SiloGOBLService) Formats(ctx context.Context, req *FetchGOBLFormats) (*GOBLFormatCollection, error) {
	p := path.Join(siloBasePath, goblPath, goblConvertPath, goblConvertFormatsPath)
	if req != nil {
		query := make(url.Values)
		if req.Country != "" {
			query.Add("country", req.Country)
		}
		if req.Schema != "" {
			query.Add("schema", req.Schema)
		}
		if req.Direction != "" {
			query.Add("direction", req.Direction)
		}
		if len(query) > 0 {
			p = p + "?" + query.Encode()
		}
	}
	col := new(GOBLFormatCollection)
	return col, svc.client.get(ctx, p, col)
}

// Export converts a GOBL document into the first of the requested formats that
// accepts it. The document and any attachments are streamed to the server as
// they are read, and the result is streamed back: close it once read.
func (svc *SiloGOBLService) Export(ctx context.Context, req *ExportGOBL) (*SiloExport, error) {
	if len(req.Formats) == 0 {
		return nil, errors.New("missing formats")
	}
	if req.Data == nil {
		return nil, errors.New("missing data")
	}
	for _, a := range req.Attachments {
		if a.Name == "" {
			return nil, errors.New("missing attachment name")
		}
		if a.Data == nil {
			return nil, fmt.Errorf("missing data for attachment %s", a.Name)
		}
	}

	p := path.Join(siloBasePath, goblPath, goblConvertPath, goblConvertExportPath) + "?" + formatsQuery(req.Formats).Encode()
	r := svc.client.conn.R().SetContext(ctx)
	if len(req.Attachments) == 0 {
		r.SetHeader("Content-Type", MIMEApplicationJSON).SetBody(req.Data)
		return sendExport(r, resty.MethodPost, p)
	}

	// Streamed through a pipe, so no file is held in memory. Closing the
	// reader once the request is done stops the writer if the server gave
	// up before reading it all.
	pr, pw := io.Pipe()
	defer pr.Close() //nolint:errcheck
	mw := multipart.NewWriter(pw)
	go func() {
		pw.CloseWithError(writeExportForm(mw, req))
	}()
	r.SetHeader("Content-Type", mw.FormDataContentType()).SetBody(pr)
	return sendExport(r, resty.MethodPost, p)
}

// writeExportForm writes the GOBL data part first, as the server expects,
// followed by each attachment.
func writeExportForm(mw *multipart.Writer, req *ExportGOBL) error {
	if err := writeFormPart(mw, formFileHeader("data", "data.json", MIMEApplicationJSON), "data", req.Data); err != nil {
		return err
	}
	for _, a := range req.Attachments {
		h := formFileHeader("attachment", a.Name, a.MIME)
		// The standard MIME part headers for both. Encoded-words keep text
		// that is not plain ASCII intact, and line breaks out of the header.
		if a.ID != "" {
			h.Set("Content-ID", mime.QEncoding.Encode("utf-8", a.ID))
		}
		if a.Desc != "" {
			h.Set("Content-Description", mime.QEncoding.Encode("utf-8", a.Desc))
		}
		if err := writeFormPart(mw, h, a.Name, a.Data); err != nil {
			return err
		}
	}
	return mw.Close()
}

func formFileHeader(field, name, mimeType string) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	// FormatMediaType quotes the parameters and encodes names that are not
	// plain ASCII.
	h.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": field, "filename": name}))
	h.Set("Content-Type", rawMIME(mimeType))
	return h
}

func writeFormPart(mw *multipart.Writer, h textproto.MIMEHeader, name string, data io.Reader) error {
	w, err := mw.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, data); err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}
	return nil
}

// Import converts a document in an external format, such as UBL or CII, into
// a GOBL envelope. The document is streamed to the server as it is read.
func (svc *SiloGOBLService) Import(ctx context.Context, req *ImportGOBL) (*GOBLImport, error) {
	if req.Data == nil {
		return nil, errors.New("missing data")
	}
	p := path.Join(siloBasePath, goblPath, goblConvertPath, goblConvertImportPath)
	if q := formatsQuery(req.Formats); len(q) > 0 {
		p = p + "?" + q.Encode()
	}
	out := new(GOBLImport)
	return out, svc.client.postRaw(ctx, p, rawMIME(req.MIME), req.Data, out)
}

// Envelope provides the imported data as a GOBL envelope.
func (gi *GOBLImport) Envelope() (*gobl.Envelope, error) {
	env := new(gobl.Envelope)
	if err := json.Unmarshal(gi.Data, env); err != nil {
		return nil, err
	}
	return env, nil
}

func formatsQuery(formats []string) url.Values {
	q := make(url.Values)
	for _, f := range formats {
		q.Add("format", f)
	}
	return q
}

// sendExport performs an export request, leaving the response body to be
// streamed by the caller.
func sendExport(r *resty.Request, method, p string) (*SiloExport, error) {
	res, err := r.SetDoNotParseResponse(true).Execute(method, p)
	if err != nil {
		return nil, err
	}
	if err := rawResponseError(res); err != nil {
		return nil, err
	}
	name := ""
	if _, params, err := mime.ParseMediaType(res.Header().Get("Content-Disposition")); err == nil {
		name = params["filename"]
	}
	return &SiloExport{
		Download: Download{
			Name: name,
			Type: res.Header().Get("Content-Type"),
			Data: res.Body,
		},
		Format: res.Header().Get(ConvertFormatHeader),
	}, nil
}
