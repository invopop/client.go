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

	FetchCodeCollection() (*CodeCollection, error)
	FetchCode(codeID string) (*Code, error)
	CreateCode(params *CodeParameters) (*Code, error)
	FetchEntryCollection(codeID string) (*EntryCollection, error)
	FetchEntry(codeID string, entryID string) (*Entry, error)
	CreateEntry(codeID string, params *EntryParameters) (*Entry, error)
}

// Sequence defines a wrapper client for the API to reach the sequence
// resources.
type Sequence struct {
	baseUrl string
	apiKey  string
	client  *http.Client
	baseReq *sling.Sling
}

// New instantiates a new instance of the sequence wrapper client with a
// simple http Client
func New(url string, apiKey string) Client {
	s := new(Sequence)

	s.baseUrl = fmt.Sprintf("%s/sequence/", url) // TODO: check trailing slash
	s.apiKey = apiKey
	s.client = &http.Client{
		Timeout: time.Second * 10,
	}
	s.baseReq = sling.New().
		Base(s.baseUrl).
		Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey)).
		Client(s.client)

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

// FetchCodeCollection returns a list of code objects belonging to the user's
// owner.
func (s *Sequence) FetchCodeCollection() (*CodeCollection, error) {
	path := "codes"
	codes := new(CodeCollection)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, codes); err != nil {
		return nil, err
	}

	return codes, nil
}

// FetchCode returns a specific code for a given code identifier, if it belongs
// to the user's owner.
func (s *Sequence) FetchCode(codeID string) (*Code, error) {
	path := fmt.Sprintf("code/%s", codeID)
	code := new(Code)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, code); err != nil {
		return nil, err
	}

	return code, nil
}

// CreateCode returns a new code with the given parameters.
func (s *Sequence) CreateCode(params *CodeParameters) (*Code, error) {
	path := "code"
	code := new(Code)
	req := s.baseReq.New().Post(path).BodyJSON(params)

	if err := s.sendRequest(req, code); err != nil {
		return nil, err
	}

	return code, nil
}

// FetchEntryCollection returns a list of entry objects belonging to the user's
// code.
func (s *Sequence) FetchEntryCollection(codeID string) (*EntryCollection, error) {
	path := fmt.Sprintf("code/%s/entries", codeID)
	entries := new(EntryCollection)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// FetchEntry returns a specific entry for a given entry identifier, if it
// belongs to the user's code.
func (s *Sequence) FetchEntry(codeID string, entryID string) (*Entry, error) {
	path := fmt.Sprintf("code/%s/entry/%s", codeID, entryID)
	entry := new(Entry)
	req := s.baseReq.New().Get(path)

	if err := s.sendRequest(req, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// CreateEntry returns a new entry with the given parameters, and asigning it
// to code's entries.
func (s *Sequence) CreateEntry(codeID string, params *EntryParameters) (*Entry, error) {
	path := fmt.Sprintf("code/%s/entry", codeID)
	entry := new(Entry)
	req := s.baseReq.New().Post(path).BodyJSON(params)

	if err := s.sendRequest(req, entry); err != nil {
		return nil, err
	}

	return entry, nil
}
