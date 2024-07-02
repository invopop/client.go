package invopop

import (
	"context"
	"path"
)

const (
	companyPath = "/company"
)

// CompanyService is used to access the company whose credentials
// we're using to authenticate with the API.
type CompanyService service

// Company represents a company or workspace in the system.
type Company struct {
	ID        string `json:"id" title:"ID" description:"UUID of the company." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	CreatedAt string `json:"created_at" title:"Created At" description:"The date and time the company was created." example:"2018-01-01T00:00:00.000Z"`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"The date and time the company was last updated." example:"2018-01-01T00:00:00.000Z"`

	Name    string `json:"name" title:"Name" description:"The name of the company." example:"My Company"`
	Country string `json:"country,omitempty" title:"Country" description:"The country the company is based in." example:"US"`
	Slug    string `json:"slug" title:"Slug" description:"A unique identifier for the company." example:"my-company"`
}

// Fetch will attempt to retrieve the company associated with the current
// authentication token.
func (s *CompanyService) Fetch(ctx context.Context) (*Company, error) {
	p := path.Join(accessBasePath, companyPath)
	c := new(Company)
	return c, s.client.get(ctx, p, c)
}
