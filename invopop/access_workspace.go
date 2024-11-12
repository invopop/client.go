package invopop

import (
	"context"
	"path"
)

const (
	workspacePath = "/workspace"
)

// WorkspaceService is used to access the workspace whose credentials
// we're using to authenticate with the API.
type WorkspaceService service

// Workspace represents a workspace previously known as a "company" in the system.
type Workspace struct {
	ID        string `json:"id" title:"ID" description:"UUID of the worksapce." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	CreatedAt string `json:"created_at" title:"Created At" description:"The date and time the workspace was created." example:"2018-01-01T00:00:00.000Z"`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"The date and time the workspace was last updated." example:"2018-01-01T00:00:00.000Z"`

	Name    string `json:"name" title:"Name" description:"The name of the workspace." example:"My Company"`
	Country string `json:"country,omitempty" title:"Country" description:"The country the workspace is based in." example:"US"`
	Slug    string `json:"slug" title:"Slug" description:"A unique identifier for the workspace." example:"my-company"`
	Sandbox bool   `json:"sandbox" title:"Sandbox" description:"Indicates if the workspace is in a sandbox environment." example:"true"`
}

// Fetch will attempt to retrieve the company associated with the current
// authentication token.
func (s *WorkspaceService) Fetch(ctx context.Context) (*Workspace, error) {
	p := path.Join(accessBasePath, workspacePath)
	c := new(Workspace)
	return c, s.client.get(ctx, p, c)
}
