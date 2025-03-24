package invopop

import (
	"context"
	"path"
)

const (
	orgsPath = "/orgs"
)

// OrgService is used to access the organizations associated with a user.
type OrgService service

// An Org is a top-level entity that groups together multiple workspaces
// and user positions.
type Org struct {
	ID        string `json:"id" title:"ID" description:"UUID of the organization." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	CreatedAt string `json:"created_at" title:"Created At" description:"The date and time the org was created." example:"2018-01-01T00:00:00.000Z"`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"The date and time the org was last updated." example:"2018-01-01T00:00:00.000Z"`

	Name   string `json:"name" title:"Name" description:"The name of the organization." example:"My Organization"`
	Domain string `json:"domain,omitempty" title:"Domain" description:"The domain of the organization." example:"myorg.com"`

	// Optional list of workspaces
	Workspaces []*Workspace `json:"workspaces,omitempty" title:"Workspaces" description:"Workspaces associated with the organization, if requested."`
}

// Fetch will attempt to retrieve the organizations associated with a user
// This method will only work with an oauth token
func (s *OrgService) Fetch(ctx context.Context) ([]*Org, error) {
	p := path.Join(accessBasePath, orgsPath)
	var orgs []*Org
	return orgs, s.client.get(ctx, p, &orgs)
}
