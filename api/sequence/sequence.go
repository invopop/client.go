package sequence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	api "github.com/invopop/client/api"
)

// Sequence defines an API client structure to manage sequence resources.
type Sequence struct {
	baseUrl string
	client  *http.Client
}

// New instantiates a new instance of the sequence client with its multiple
// resource's clients
func New(url string) *Sequence {
	s := new(Sequence)

	s.baseUrl = fmt.Sprintf("%s/sequence", url)
	s.client = &http.Client{
		Timeout: time.Second * 10,
	}

	return s
}

func (s *Sequence) sendRequest(req *http.Request, v interface{}) *api.ClientError {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := s.client.Do(req)
	if err != nil {
		return api.NewInternalError(err.Error())
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes api.APIError

		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err == nil {
			return api.NewError(res.StatusCode, errRes.Message)
		}

		return api.NewError(res.StatusCode, err.Error())
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return api.NewInternalError(err.Error())
	}

	return nil
}

func (s *Sequence) FetchCodeCollection(
	c context.Context,
	ownerID string,
) (*CodeCollection, *api.ClientError) {
	path := fmt.Sprintf("%s/%s/codes", s.baseUrl, ownerID)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(CodeCollection)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Sequence) FetchCode(
	c context.Context,
	ownerID string,
	codeID string,
) (*Code, *api.ClientError) {
	path := fmt.Sprintf("%s/%s/code/%s", s.baseUrl, ownerID, codeID)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(Code)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Sequence) CreateCode(
	c context.Context,
	ownerID string,
	params *CodeParameters,
) (*Code, error) {
	path := fmt.Sprintf("%s/%s/code", s.baseUrl, ownerID)

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(params)

	req, err := http.NewRequest("POST", path, payload)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(Code)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Sequence) FetchEntryCollection(
	c context.Context,
	ownerID string,
	codeID string,
) (*EntryCollection, *api.ClientError) {
	path := fmt.Sprintf("%s/%s/code/%s/entries", s.baseUrl, ownerID, codeID)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(EntryCollection)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Sequence) FetchEntry(
	c context.Context,
	ownerID string,
	codeID string,
	entryID string,
) (*Entry, *api.ClientError) {
	path := fmt.Sprintf("%s/%s/code/%s/entry/%s", s.baseUrl, ownerID, codeID, entryID)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(Entry)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Sequence) CreateEntry(
	c context.Context,
	ownerID string,
	codeID string,
	params *EntryParameters,
) (*Entry, error) {
	path := fmt.Sprintf("%s/%s/code/%s/entry", s.baseUrl, ownerID, codeID)

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(params)

	req, err := http.NewRequest("POST", path, payload)
	if err != nil {
		return nil, api.NewInternalError(err.Error())
	}
	req = req.WithContext(c)

	res := new(Entry)
	if err := s.sendRequest(req, res); err != nil {
		return nil, err
	}

	return res, nil
}
