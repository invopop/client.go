package gateway

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Poke sends a message to the gateway indicating that we've received an
// external prompt, like a webhook, and the original task should be re-sent.
func (gw *Client) Poke(ctx context.Context, req *TaskPoke) error {
	in, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	out, err := gw.nc.RequestWithContext(ctx, SubjectTasksPoke, in)
	if err != nil {
		return err
	}
	res := new(TaskPokeResponse)
	if err := proto.Unmarshal(out.Data, res); err != nil {
		return err
	}
	if res.Err != nil {
		return res.Err
	}

	// PokeTaskResponse is empty if successful
	return nil
}
