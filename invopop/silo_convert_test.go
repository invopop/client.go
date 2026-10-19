package invopop

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

const (
	testFormat    = "ubl+peppol+bis-v3"
	testFormatCII = "cii"
	testPDFMIME   = "application/pdf"
	testExported  = "<Invoice/>"
	testGOBLData  = `{"$schema":"https://gobl.org/draft-0/bill/invoice"}`
	testExportCT  = "application/xml"
	testExportCD  = `attachment; filename="Factura F-1.xml"`
	testExportErr = `{"code":"invalid","message":"no format accepts the document"}`
)

// exportResponder replies to an export as the API does, keeping the request
// and the body it read for the test to check.
type exportResponder struct {
	t      *testing.T
	status int
	req    *http.Request
	body   []byte
}

func (r *exportResponder) respond(req *http.Request) (*http.Response, error) {
	r.req = req
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		require.NoError(r.t, err)
		r.body = b
	}
	if r.status != 0 {
		return &http.Response{
			StatusCode: r.status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(testExportErr)),
			Request:    req,
		}, nil
	}
	// Set rather than a literal, so the header names are canonical as they
	// would be off the wire.
	h := make(http.Header)
	h.Set("Content-Type", testExportCT)
	h.Set("Content-Disposition", testExportCD)
	h.Set(ConvertFormatHeader, testFormat)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(testExported)),
		Request:    req,
	}, nil
}

func exportClient(t *testing.T, r *exportResponder) *Client {
	t.Helper()
	r.t = t
	c := New()
	c.conn = resty.NewWithClient(testy.HTTPClient(r.respond)).SetBaseURL("https://api.test")
	return c
}

func assertExported(t *testing.T, out *SiloExport) {
	t.Helper()
	defer out.Close() //nolint:errcheck
	assert.Equal(t, testFormat, out.Format)
	assert.Equal(t, testExportCT, out.Type)
	assert.Equal(t, "Factura F-1.xml", out.Name)
	data, err := io.ReadAll(out.Data)
	require.NoError(t, err)
	assert.Equal(t, testExported, string(data))
}

func TestSiloGOBLFormats(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"list":[{"key":"ubl+peppol+bis-v3","mime":"application/xml","attachments":true}]}`)
	c := testClient(t, rec)

	col, err := c.Silo().GOBL().Formats(context.Background(), &FetchGOBLFormats{
		Country:   "DE",
		Direction: ConvertDirectionExport,
	})
	require.NoError(t, err)
	require.Len(t, col.List, 1)
	assert.Equal(t, testFormat, col.List[0].Key)
	assert.True(t, col.List[0].Attachments)

	req := rec.last()
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/silo/v1/gobl/convert/formats", req.URL.Path)
	assert.Equal(t, "DE", req.URL.Query().Get("country"))
	assert.Equal(t, "export", req.URL.Query().Get("direction"))
}

func TestSiloGOBLExport(t *testing.T) {
	r := new(exportResponder)
	c := exportClient(t, r)

	out, err := c.Silo().GOBL().Export(context.Background(), &ExportGOBL{
		Formats: []string{testFormat, testFormatCII},
		Data:    strings.NewReader(testGOBLData),
	})
	require.NoError(t, err)
	assertExported(t, out)

	assert.Equal(t, http.MethodPost, r.req.Method)
	assert.Equal(t, "/silo/v1/gobl/convert/export", r.req.URL.Path)
	// Escaped, so the "+" is not taken as a space.
	assert.Contains(t, r.req.URL.RawQuery, "format=ubl%2Bpeppol%2Bbis-v3")
	assert.Equal(t, []string{testFormat, testFormatCII}, r.req.URL.Query()["format"])
	// A form even without attachments, as the server reads JSON whole.
	mt, params, err := mime.ParseMediaType(r.req.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/form-data", mt)
	mr := multipart.NewReader(strings.NewReader(string(r.body)), params["boundary"])
	p, err := mr.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "data", p.FormName())
	data, err := io.ReadAll(p)
	require.NoError(t, err)
	assert.JSONEq(t, testGOBLData, string(data))
	_, err = mr.NextPart()
	assert.Equal(t, io.EOF, err)
}

func TestSiloGOBLExportWithAttachments(t *testing.T) {
	r := new(exportResponder)
	c := exportClient(t, r)

	pdf := "%PDF-1.7 receipt"
	out, err := c.Silo().GOBL().Export(context.Background(), &ExportGOBL{
		Formats: []string{testFormat},
		Data:    strings.NewReader(testGOBLData),
		Attachments: []*ExportGOBLAttachment{
			{ID: "REC-1", Name: "recibo nº 1.pdf", Desc: "Recibo nº 1\r\nX-Injected: yes", MIME: testPDFMIME, Data: strings.NewReader(pdf)},
			{Name: "notes.txt", Data: strings.NewReader("")},
		},
	})
	require.NoError(t, err)
	assertExported(t, out)

	mt, params, err := mime.ParseMediaType(r.req.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/form-data", mt)
	mr := multipart.NewReader(strings.NewReader(string(r.body)), params["boundary"])

	type part struct{ field, name, id, desc, mime, data string }
	var parts []part
	dec := new(mime.WordDecoder)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		data, err := io.ReadAll(p)
		require.NoError(t, err)
		assert.Empty(t, p.Header.Get("X-Injected"), "a line break must not start a new header")
		id, err := dec.DecodeHeader(p.Header.Get("Content-ID"))
		require.NoError(t, err)
		desc, err := dec.DecodeHeader(p.Header.Get("Content-Description"))
		require.NoError(t, err)
		parts = append(parts, part{p.FormName(), p.FileName(), id, desc, p.Header.Get("Content-Type"), string(data)})
	}
	assert.Equal(t, []part{
		{"data", "data.json", "", "", MIMEApplicationJSON, testGOBLData},
		{"attachment", "recibo nº 1.pdf", "REC-1", "Recibo nº 1\r\nX-Injected: yes", testPDFMIME, pdf},
		{"attachment", "notes.txt", "", "", defaultFileMIME, ""},
	}, parts)
}

func TestSiloGOBLExportValidation(t *testing.T) {
	c := New()
	ctx := context.Background()
	_, err := c.Silo().GOBL().Export(ctx, &ExportGOBL{Data: strings.NewReader("{}")})
	assert.EqualError(t, err, "missing formats")
	_, err = c.Silo().GOBL().Export(ctx, &ExportGOBL{Formats: []string{testFormatCII}})
	assert.EqualError(t, err, errMissingData)
	_, err = c.Silo().GOBL().Export(ctx, &ExportGOBL{
		Formats:     []string{testFormatCII},
		Data:        strings.NewReader("{}"),
		Attachments: []*ExportGOBLAttachment{{Data: strings.NewReader("x")}},
	})
	assert.EqualError(t, err, "missing attachment name")
}

func TestSiloGOBLExportError(t *testing.T) {
	r := &exportResponder{status: http.StatusUnprocessableEntity}
	c := exportClient(t, r)

	out, err := c.Silo().GOBL().Export(context.Background(), &ExportGOBL{
		Formats: []string{testFormatCII},
		Data:    strings.NewReader(testGOBLData),
	})
	require.Error(t, err)
	assert.Nil(t, out)
	re := AsResponseError(err)
	require.NotNil(t, re)
	assert.Equal(t, http.StatusUnprocessableEntity, re.StatusCode())
	// The message is read even though the body was left unparsed.
	assert.Equal(t, "no format accepts the document", re.Message)
}

func TestSiloGOBLImport(t *testing.T) {
	rec := newRecorder(http.StatusOK, `{"format":"ubl+peppol+bis-v3","data":{"$schema":"https://gobl.org/draft-0/envelope"},"attachments":[{"id":"att-1","name":"receipt.pdf","mime":"application/pdf","size":8,"data":"JVBERi0xLjc="}]}`)
	c := testClient(t, rec)

	out, err := c.Silo().GOBL().Import(context.Background(), &ImportGOBL{
		MIME: "application/xml",
		Data: strings.NewReader(testExported),
	})
	require.NoError(t, err)
	assert.Equal(t, testFormat, out.Format)
	assert.JSONEq(t, `{"$schema":"https://gobl.org/draft-0/envelope"}`, string(out.Data))
	require.Len(t, out.Attachments, 1)
	assert.Equal(t, "%PDF-1.7", string(out.Attachments[0].Data))

	req := rec.last()
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/silo/v1/gobl/convert/import", req.URL.Path)
	assert.Empty(t, req.URL.RawQuery)
	assert.Equal(t, "application/xml", req.Header.Get("Content-Type"))
	assert.Equal(t, testExported, rec.lastBody())
}

func TestSiloEntriesExport(t *testing.T) {
	r := new(exportResponder)
	c := exportClient(t, r)

	out, err := c.Silo().Entries().Export(context.Background(), &ExportSiloEntry{
		ID:          testEntryID,
		Formats:     []string{testFormat},
		Attachments: ExportAttachmentsAttachment,
		MIME:        []string{testPDFMIME},
	})
	require.NoError(t, err)
	assertExported(t, out)

	assert.Equal(t, http.MethodGet, r.req.Method)
	assert.Equal(t, "/silo/v1/entries/"+testEntryID+"/export", r.req.URL.Path)
	q := r.req.URL.Query()
	assert.Equal(t, testFormat, q.Get("format"))
	assert.Equal(t, "attachment", q.Get("attachments"))
	assert.Equal(t, testPDFMIME, q.Get("mime"))
}

func TestSiloEntriesExportValidation(t *testing.T) {
	c := New()
	_, err := c.Silo().Entries().Export(context.Background(), &ExportSiloEntry{Formats: []string{testFormatCII}})
	assert.EqualError(t, err, "missing id")
	_, err = c.Silo().Entries().Export(context.Background(), &ExportSiloEntry{ID: testEntryID})
	assert.EqualError(t, err, "missing formats")
}

func TestSiloEntriesImport(t *testing.T) {
	t.Run("posts without an ID", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1"}`)
		c := testClient(t, rec)
		e, err := c.Silo().Entries().Import(context.Background(), &ImportSiloEntry{
			Key:  "inv-1",
			Data: strings.NewReader(testExported),
		})
		require.NoError(t, err)
		assert.Equal(t, testEntryID, e.ID)

		req := rec.last()
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/silo/v1/entries/import", req.URL.Path)
		assert.Equal(t, "key=inv-1", req.URL.RawQuery)
		assert.Equal(t, defaultFileMIME, req.Header.Get("Content-Type"))
		assert.Equal(t, testExported, rec.lastBody())
	})

	t.Run("puts with an ID and passes every option", func(t *testing.T) {
		rec := newRecorder(http.StatusOK, `{"id":"entry-1"}`)
		c := testClient(t, rec)
		_, err := c.Silo().Entries().Import(context.Background(), &ImportSiloEntry{
			ID:           testEntryID,
			Key:          "inv-1",
			Folder:       testFolder,
			Formats:      []string{testFormat},
			Name:         "invoice.xml",
			AllowInvalid: true,
			Sign:         true,
			Private:      true,
			MIME:         "application/xml",
			Data:         strings.NewReader(testExported),
		})
		require.NoError(t, err)

		req := rec.last()
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/silo/v1/entries/"+testEntryID+"/import", req.URL.Path)
		q := req.URL.Query()
		assert.Equal(t, "inv-1", q.Get("key"))
		assert.Equal(t, testFolder, q.Get("folder"))
		assert.Equal(t, testFormat, q.Get("format"))
		assert.Equal(t, "invoice.xml", q.Get("name"))
		assert.Equal(t, "true", q.Get("allow_invalid"))
		assert.Equal(t, "true", q.Get("sign"))
		assert.Equal(t, "true", q.Get("private"))
		assert.Equal(t, "application/xml", req.Header.Get("Content-Type"))
	})

	t.Run("requires data", func(t *testing.T) {
		_, err := New().Silo().Entries().Import(context.Background(), &ImportSiloEntry{})
		assert.EqualError(t, err, errMissingData)
	})
}
