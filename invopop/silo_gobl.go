package invopop

import (
	"context"
	"encoding/json"
	"path"
)

const (
	goblPath      = "gobl"
	goblBuildPath = "build"
	goblSignPath  = "sign"
)

// SiloGOBLService provides access to GOBL build and sign endpoints useful for ensuring
// documents are valid without worrying about versions of local libraries.
//
// Requests to this endpoint are authenticated, but are stateless and thus fast.
type SiloGOBLService service

// GOBL contains the response from the Silo GOBL service with the contents.
type GOBL struct {
	Data json.RawMessage `json:"data"`
}

// BuildGOBL defines the fields required to build a GOBL object.
type BuildGOBL struct {
	Data    json.RawMessage `json:"data" title:"Data" description:"GOBL Envelope or Object to calculate and validate."`
	Envelop bool            `json:"envelop" title:"Envelop" description:"When true, a complete GOBL Envelope will be provided as opposed to the standalone object."`
}

// SignGOBL defines the fields required to sign a GOBL object.
type SignGOBL struct {
	Data json.RawMessage `json:"data" title:"Data" description:"GOBL Envelope or Object to sign."`
}

// Build sends a request to build the input document
func (svc *SiloGOBLService) Build(ctx context.Context, in *BuildGOBL) (*GOBL, error) {
	out := new(GOBL)
	return out, svc.client.post(ctx, path.Join(siloBasePath, goblPath, goblBuildPath), in, out)
}

// Sign sends a request to sign the input document
func (svc *SiloGOBLService) Sign(ctx context.Context, in *SignGOBL) (*GOBL, error) {
	out := new(GOBL)
	return out, svc.client.post(ctx, path.Join(siloBasePath, goblPath, goblSignPath), in, out)
}
