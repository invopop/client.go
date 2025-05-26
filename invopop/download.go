package invopop

import (
	"context"
	"io"
	"mime"
	"path"
	"strings"

	"resty.dev/v3"
)

// Download contains a generic structure to contain response data
// from a file download.
type Download struct {
	Name string
	Type string
	Data io.ReadCloser
}

// Close the download
func (d *Download) Close() error {
	if d == nil || d.Data == nil {
		return nil
	}
	return d.Data.Close()
}

const (
	spoolProtocol string = "spool:"
)

// Download will look at the complete URL and determine if it should
// request the file directly, or re-write the URL to download from
// an Invopop endpoint using the current credentials.
// NOTE: please close the Download response after reading to avoid
// memory leaks.
func (c *Client) Download(ctx context.Context, url string) (*Download, error) {
	// Detect known special URLs
	if key, ok := strings.CutPrefix(url, spoolProtocol); ok {
		return c.Silo().Spool().Download(ctx, key)
	}

	// Perform a regular download
	res, err := resty.New().R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return nil, err
	}
	d := &Download{
		Name: extractContentFilename(res, url),
		Type: res.Header().Get("Content-Type"),
		Data: res.Body,
	}
	return d, nil
}

func extractContentFilename(res *resty.Response, url string) string {
	_, params, err := mime.ParseMediaType(res.Header().Get("Content-Disposition"))
	filename := ""
	if err == nil && params["filename"] != "" {
		filename = params["filename"]
	} else {
		// Fallback to URL path if header not available
		filename = path.Base(url)
	}
	return filename
}
