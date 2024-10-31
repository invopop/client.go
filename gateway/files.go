package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/protobuf/proto"
)

// CreateFile allows us to build a file place holder and upload the data afterwards
// by posting to the URL provided.
func (gw *Client) CreateFile(ctx context.Context, req *CreateFile) (*File, error) {
	in, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	out, err := gw.nc.RequestWithContext(ctx, SubjectFilesCreate, in)
	if err != nil {
		return nil, err
	}
	res := new(FileResponse)
	if err := proto.Unmarshal(out.Data, res); err != nil {
		return nil, err
	}
	if res.Err != nil {
		return nil, res.Err
	}
	return res.File, nil
}

// CreateAndUploadFile makes it easier to upload a file using basic details such
// as the ID, Job, Envelope, Name, and Description, and automatically add the
// SHA256, MIME, and Size attributes based on the provided data.
// The tradeoff here as opposed to two separate calls is that the data is kept
// in memory and not sent through a buffer.
func (gw *Client) CreateAndUploadFile(ctx context.Context, req *CreateFile, data []byte) (*File, error) {
	gw.prepareCreateFileFromData(req, data)

	f, err := gw.CreateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)
	if err := gw.UploadFile(ctx, f, r); err != nil {
		return nil, err
	}

	return f, nil
}

// UploadFile performs an HTTP PUT action to send the data to the silo. The URL comes
// from the file object. Provided the SHA256 data matches, the file uploaded will
// function as expected.
func (gw *Client) UploadFile(ctx context.Context, f *File, data io.Reader) error {
	url, err := gw.fileUploadURL(f)
	if err != nil {
		return fmt.Errorf("upload url: %w", err)
	}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", f.Mime)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("upload error, status: %s", res.Status)
	}
	return nil
}

// FetchFile performs an HTTP GET action to retrieve a file's data from the
// silo. The URL comes from the file object.
func (gw *Client) FetchFile(ctx context.Context, f *File) ([]byte, error) {
	url, err := gw.fileUploadURL(f)
	if err != nil {
		return nil, fmt.Errorf("upload url: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	req = req.WithContext(ctx)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("fetch error, status: %s", res.Status)
	}
	return io.ReadAll(res.Body)
}

func (gw *Client) prepareCreateFileFromData(req *CreateFile, data []byte) {
	req.Size = int32(len(data))
	if req.Mime == "" {
		mt := mimetype.Detect(data)
		req.Mime = mt.String()
	}
	sum := sha256.Sum256(data)
	req.Sha256 = hex.EncodeToString(sum[:])
}

// fileUploadURL determines the URL used to upload to. We don't use the URL
// provided by the silo service as that is only reliable for exterior public
// use.
func (gw *Client) fileUploadURL(f *File) (string, error) {
	if gw.siloPublicBaseURL == "" {
		return "", errors.New("missing silo public base url")
	}
	return fmt.Sprintf("%s/%s/%s/%s?h=%s",
		gw.siloPublicBaseURL,
		f.SiloEntryId,
		f.Id,
		f.Name,
		f.Hash,
	), nil
}
