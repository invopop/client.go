package sequence

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	api "github.com/invopop/client/api"
)

// Sequence defines a wrapper client for the API to reach the sequence
// resources.
type Sequence struct {
	baseUrl string
	client  *http.Client
	baseReq *sling.Sling
}

// New instantiates a new instance of the sequence wrapper client with a
// simple http Client
func New(url string) *Sequence {
	s := new(Sequence)

	s.baseUrl = fmt.Sprintf("%s/sequence/", url) // TODO: check trailing slash
	s.client = &http.Client{
		Timeout: time.Second * 10,
	}
	s.baseReq = sling.New().Base(s.baseUrl).Client(s.client)

	return s
}

func (s *Sequence) sendRequest(req *sling.Sling, succV interface{}) *api.ClientError {
	failV := new(api.APIError)

	resp, err := req.Receive(succV, failV)
	if err != nil {
		return api.NewInternalError(err.Error())
	}

	if !failV.IsNil() {
		return api.NewError(resp.StatusCode, failV.Message)
	}

	return nil
}

func (s *Sequence) FetchCodeCollection(
	ownerID string,
) (*CodeCollection, *api.ClientError) {
	path := fmt.Sprintf("%s/codes", ownerID)
	codes := new(CodeCollection)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, codes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *Sequence) FetchCode(
	ownerID string,
	codeID string,
) (*Code, *api.ClientError) {
	path := fmt.Sprintf("%s/code/%s", ownerID, codeID)
	code := new(Code)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, code); err != nil {
		return nil, err
	}

	return code, nil
}

func (s *Sequence) CreateCode(
	ownerID string,
	params *CodeParameters,
) (*Code, *api.ClientError) {
	path := fmt.Sprintf("%s/code", ownerID)
	code := new(Code)
	req := s.baseReq.New().Post(path).BodyJSON(params)

	if err := s.sendRequest(req, code); err != nil {
		return nil, err
	}

	return code, nil
}

func (s *Sequence) FetchEntryCollection(
	ownerID string,
	codeID string,
) (*EntryCollection, *api.ClientError) {
	path := fmt.Sprintf("%s/code/%s/entries", ownerID, codeID)
	entries := new(EntryCollection)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (s *Sequence) FetchEntry(
	ownerID string,
	codeID string,
	entryID string,
) (*Entry, *api.ClientError) {
	path := fmt.Sprintf("%s/code/%s/entry/%s", ownerID, codeID, entryID)
	entry := new(Entry)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Sequence) CreateEntry(
	ownerID string,
	codeID string,
	params *EntryParameters,
) (*Entry, *api.ClientError) {
	path := fmt.Sprintf("%s/code/%s/entry", ownerID, codeID)
	entry := new(Entry)
	req := s.baseReq.New().Post(path).BodyJSON(params)

	if err := s.sendRequest(req, entry); err != nil {
		return nil, err
	}

	return entry, nil
}
