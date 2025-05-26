package invopop

import (
	"context"
	"errors"
	"path"
)

const (
	spoolPath = "spool"
)

// SiloSpoolService provides access to the spool endpoints. Upload requires
// the "enrolled" scope, but download and delete can be performed using
// regular credentials.
type SiloSpoolService service

type siloSpoolUpload struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type siloSpoolUploadResponse struct {
	Key string `json:"key"`
}

// Upload sends the file data to the silo spool service for storage, and returns
// a "key" that can be used for retrieval.
func (s *SiloSpoolService) Upload(ctx context.Context, name, mediaType string, data []byte) (string, error) {
	if name == "" {
		return "", errors.New("missing name")
	}
	if mediaType == "" {
		return "", errors.New("missing media type")
	}
	if len(data) == 0 {
		return "", errors.New("missing data")
	}
	req := &siloSpoolUpload{
		Name: name,
		Type: mediaType,
		Data: data,
	}
	p := path.Join(siloBasePath, spoolPath)
	out := new(siloSpoolUploadResponse)
	return out.Key, s.client.post(ctx, p, req, out)
}

// Download will try to download the object from the silo spool identified
// by the key.
func (s *SiloSpoolService) Download(ctx context.Context, key string) (*Download, error) {
	if key == "" {
		return nil, errors.New("missing key")
	}
	p := path.Join(siloBasePath, spoolPath, key)
	re := new(ResponseError)
	res, err := s.client.conn.R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		SetError(re).
		Get(p)
	if err != nil {
		return nil, err
	}
	if err := re.handle(res); err != nil {
		return nil, err
	}
	d := &Download{
		Name: extractContentFilename(res, p),
		Type: res.Header().Get("Content-Type"),
		Data: res.Body,
	}
	return d, nil
}

// Delete sends a request to delete the silo spool object by key.
func (s *SiloSpoolService) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("missing key")
	}
	p := path.Join(siloBasePath, spoolPath, key)
	return s.client.delete(ctx, p, nil)
}
