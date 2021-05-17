package sequence

import (
	"context"
	"time"

	"github.com/google/uuid"
	keystore "github.com/invopop/keystore/pkg/keystore"
	sequence "github.com/invopop/sequence/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SequenceClient defines a client structure to manage the resources of the
// sequence service.
type SequenceClient struct {
	signer  *keystore.Signer
	code    sequence.CodeResourceClient
	codes   sequence.CodesResourceClient
	entry   sequence.EntryResourceClient
	entries sequence.EntriesResourceClient
}

// New instantiates a new instance of the sequence client with its multiple
// resource clients
func New(signer *keystore.Signer, grpc *grpc.ClientConn) *SequenceClient {
	cli := new(SequenceClient)

	cli.signer = signer
	cli.code = sequence.NewCodeResourceClient(grpc)
	cli.codes = sequence.NewCodesResourceClient(grpc)
	cli.entry = sequence.NewEntryResourceClient(grpc)
	cli.entries = sequence.NewEntriesResourceClient(grpc)

	return cli
}

func (cli *SequenceClient) FetchCodeCollection(ownerID string) (*sequence.CodesResponse, error) {
	var res = new(sequence.CodesResponse)

	req := &sequence.FetchCodesRequest{
		Owner: &sequence.Owner{
			Id: ownerID,
		},
	}

	codesRes, err := cli.codes.Fetch(context.Background(), req)
	if err != nil {
		return nil, err
	}

	res.Codes = codesRes.Codes

	return res, nil
}

func (cli *SequenceClient) FetchCode(ownerID string, codeID string) (*sequence.CodeResponse, error) {
	return cli.validatedFetchCode(ownerID, codeID)
}

func (cli *SequenceClient) CreateCode(ownerID string, params *CodeParameters) (*sequence.CodeResponse, error) {
	var err error

	newCodeReq := sequence.CreateCodeRequest{
		Id: uuid.New().String(),
		Owner: &sequence.Owner{
			Id:   ownerID,
			Name: params.Name,
		},
		Name:    params.Name,
		Prefix:  params.Prefix,
		Suffix:  params.Suffix,
		Padding: params.Padding,
	}

	codeRes, err := cli.code.Create(context.Background(), &newCodeReq)
	if err != nil {
		return nil, err
	}

	return codeRes, nil
}

func (cli *SequenceClient) FetchEntryCollection(ownerID string, codeID string) (*sequence.EntriesResponse, error) {
	if _, err := cli.validatedFetchCode(ownerID, codeID); err != nil {
		return nil, err
	}

	req := &sequence.FetchEntriesRequest{
		Owner: &sequence.Owner{
			Id: ownerID,
		},
		CodeId: codeID,
	}

	entriesRes, err := cli.entries.Fetch(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return entriesRes, nil
}

func (cli *SequenceClient) FetchEntry(ownerID string, codeID string, entryID string) (*sequence.EntryResponse, error) {
	var err error

	codeRes, err := cli.validatedFetchCode(ownerID, codeID)
	if err != nil {
		return nil, err
	}

	fetchEntryReq := sequence.FetchEntryRequest{
		Id: entryID,
		Owner: &sequence.Owner{
			Id: ownerID,
		},
	}

	entryRes, err := cli.entry.Fetch(context.Background(), &fetchEntryReq)
	if err != nil {
		return nil, err
	}

	if codeRes.Code.Id != entryRes.Entry.CodeId {
		return nil, status.Error(codes.NotFound, "code id mismatch")
	}

	return entryRes, nil
}

func (cli *SequenceClient) CreateEntry(ownerID string, codeID string, params *EntryParameters) (*sequence.EntryResponse, error) {
	var err error

	if _, err := cli.validatedFetchCode(ownerID, codeID); err != nil {
		return nil, err
	}

	newEntryReq := sequence.CreateEntryRequest{
		Id:     uuid.New().String(),
		CodeId: codeID,
		Meta:   params.Meta,
	}

	sig, err := cli.signer.Sign(sequence.EntrySigMsg{
		Id:   newEntryReq.Id,
		Cid:  newEntryReq.CodeId,
		Pid:  "",
		Num:  "",
		Meta: params.Meta,
		Ts:   time.Now().Unix(),
	})

	if err != nil {
		return nil, err
	}

	newEntryReq.Sig = sig.String()

	entryRes, err := cli.entry.Create(context.Background(), &newEntryReq)
	if err != nil {
		return nil, err
	}

	return entryRes, nil
}
