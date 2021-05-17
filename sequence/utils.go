package sequence

import (
	"context"

	sequence "github.com/invopop/sequence/protocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cli *SequenceClient) validatedFetchCode(ownerID string, codeID string) (*sequence.CodeResponse, error) {
	req := &sequence.FetchCodeRequest{
		Id: codeID,
	}

	codeRes, err := cli.code.Fetch(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if ownerID != codeRes.Code.Owner.Id {
		return nil, status.Error(codes.NotFound, "owner id mismatch")
	}

	return codeRes, nil
}
