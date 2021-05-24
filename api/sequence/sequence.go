package sequence

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	api "github.com/invopop/client/api"
)

// Client defines a the key methods to access the resources of the sequence
// service.
type Client interface {
	// sendRequest is a general purpose method to make requests
	sendRequest(req *sling.Sling, succV interface{}) error

	FetchCodeCollection(ownerID string) (*CodeCollection, error)
	FetchCode(ownerID string, codeID string) (*Code, error)
	CreateCode(ownerID string, params *CodeParameters) (*Code, error)
	FetchEntryCollection(ownerID string, codeID string) (*EntryCollection, error)
	FetchEntry(ownerID string, codeID string, entryID string) (*Entry, error)
	CreateEntry(ownerID string, codeID string, params *EntryParameters) (*Entry, error)
}

// Sequence defines a wrapper client for the API to reach the sequence
// resources.
type Sequence struct {
	baseUrl string
	client  *http.Client
	baseReq *sling.Sling
}

// New instantiates a new instance of the sequence wrapper client with a
// simple http Client
func New(url string) Client {
	s := new(Sequence)

	s.baseUrl = fmt.Sprintf("%s/sequence/", url) // TODO: check trailing slash
	s.client = &http.Client{
		Timeout: time.Second * 10,
	}
	s.baseReq = sling.New().Base(s.baseUrl).Client(s.client)

	return s
}

func (s *Sequence) sendRequest(req *sling.Sling, succV interface{}) error {
	failV := new(api.APIError)

	resp, err := req.Receive(succV, failV)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error())
	}

	if failV.Message != "" {
		return api.NewError(resp.StatusCode, failV.Message)
	}

	return nil
}

func (s *Sequence) FetchCodeCollection(
	ownerID string,
) (*CodeCollection, error) {
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
) (*Code, error) {
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
) (*Code, error) {
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
) (*EntryCollection, error) {
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
) (*Entry, error) {
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
) (*Entry, error) {
	path := fmt.Sprintf("%s/code/%s/entry", ownerID, codeID)
	entry := new(Entry)
	req := s.baseReq.New().Post(path).BodyJSON(params)

	if err := s.sendRequest(req, entry); err != nil {
		return nil, err
	}

	return entry, nil
}
