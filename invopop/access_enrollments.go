package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/invopop/gobl/uuid"
)

const (
	enrollmentPath = "/enrollment"
	authorizePath  = "authorize"
)

// EnrollmentService helps manage access to a single enrollment
// that is associated between the app and the "owner" of the workspace.
type EnrollmentService service

// Enrollment represents an enrollment in the system.
type Enrollment struct {
	ID        string `json:"id" title:"ID" description:"UUID of the enrollment." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	CreatedAt string `json:"created_at" title:"Created At" description:"The date and time the enrollment was created." example:"2018-01-01T00:00:00.000Z"`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"The date and time the enrollment was last updated." example:"2018-01-01T00:00:00.000Z"`

	OwnerID string `json:"owner_id" title:"Owner ID" description:"The ID of the entity that owns the enrollment." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	AppID   string `json:"app_id" title:"Application ID" description:"ID of the application associated with the enrollment." example:"01900e17-db4d-78a5-8505-c93ae63e8a0d"`
	Sandbox bool   `json:"sandbox" title:"Sandbox" description:"Indicates if the enrollment's workspace is in a sandbox environment." example:"false"`

	Data json.RawMessage `json:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`

	Disabled bool `json:"disabled" title:"Disabled" description:"Whether the enrollment is disabled." example:"false"`

	Token string `json:"token" title:"Token" description:"A token that may be used to authenticate the enrollment with API operations."`
}

// authorizeEnrollment is used internally to describe the fields required to confirm
// that an app has access to the enrollment details for the Owner.
type authorizeEnrollment struct {
	ID           string `json:"id,omitempty" title:"ID" description:"Enrollment ID to use when the owner ID is not available" example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	OwnerID      string `json:"owner_id,omitempty" title:"Owner ID" description:"The ID of the entity that owns the enrollment. It is essential this is provided from a trusted source or an auth token is provided in the headers." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	ClientID     string `json:"client_id" title:"Client ID" description:"The ID of the application that is being enrolled." example:"CvI6CIygjGP10g"`
	ClientSecret string `json:"client_secret" title:"Client Secret" description:"The secret key of the application that is being enrolled." example:"YSKIfGaUrEdDFK_NPGO-Yj1oVDJcjV15N4hHbuAEg2c"`
}

// CreateEnrollment is used by apps to create an enrollment on behalf of
// an end user after choosing a workspace.
type createEnrollment struct {
	ID           string          `param:"id" title:"ID" description:"UUIDv7 of the new enrollment to create." example:"01950020-daef-7d75-b1ba-33e7e392a658"`
	OwnerID      string          `json:"owner_id" title:"Owner ID" description:"Workspace ID to associate with the enrollment."`
	ClientID     string          `json:"client_id" title:"Client ID" description:"The ID of the application that is being enrolled." example:"XzhLPeXCi3GBVg"`
	ClientSecret string          `json:"client_secret" title:"Client Secret" description:"The secret key of the application that is being enrolled." example:"p2NWtVpuDxDYt41crWUBmQKaE4Mh92roDxp_8UKkIJY"`
	Data         json.RawMessage `json:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`
}

// UpdateEnrollment defines the request body for updating an enrollment.
type UpdateEnrollment struct {
	Data json.RawMessage `param:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`
}

// Authorize tries to update the Enrollment object with an embedded token to use
// for subsequent requests to the API.
//
// OAuth credentials must have been configured in the client for this to work, and
// will be used alongside regular token authentication to ensure the client has
// the necessary permissions to generate the enrollment token.
func (s *EnrollmentService) Authorize(ctx context.Context) (*Enrollment, error) {
	return s.AuthorizeWithOwnerID(ctx, "")
}

// AuthorizeWithID will make a request to load the enrollment using the app credentials
// and a specific enrollment ID. This is useful when the owner ID is not available.
func (s *EnrollmentService) AuthorizeWithID(ctx context.Context, id string) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath, authorizePath)
	if s.client.clientID == "" || s.client.clientSecret == "" {
		return nil, errors.New("missing OAuth client credentials: client_id and client_secret")
	}
	req := &authorizeEnrollment{
		ID:           id,
		ClientID:     s.client.clientID,
		ClientSecret: s.client.clientSecret,
	}
	e := new(Enrollment)
	return e, s.client.post(ctx, p, req, e)
}

// AuthorizeWithOwnerID allows applications to authorize an enrollment with a specific owner ID.
func (s *EnrollmentService) AuthorizeWithOwnerID(ctx context.Context, ownerID string) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath, authorizePath)
	if s.client.clientID == "" || s.client.clientSecret == "" {
		return nil, errors.New("missing OAuth client credentials: client_id and client_secret")
	}
	req := &authorizeEnrollment{
		OwnerID:      ownerID, // may be empty if intention is to use token
		ClientID:     s.client.clientID,
		ClientSecret: s.client.clientSecret,
	}
	e := new(Enrollment)
	return e, s.client.post(ctx, p, req, e)
}

// Fetch retrieves an enrollment associated with the current authentication token.
func (s *EnrollmentService) Fetch(ctx context.Context) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath)
	e := new(Enrollment)
	return e, s.client.get(ctx, p, e)
}

// Update allows applications to update their enrollment's embedded data.
func (s *EnrollmentService) Update(ctx context.Context, req *UpdateEnrollment) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath)
	e := new(Enrollment)
	return e, s.client.post(ctx, p, req, e)
}

// Create will create an enrollment between a workspace and an application.
func (s *EnrollmentService) Create(ctx context.Context, ownerID string) (*Enrollment, error) {
	enrollmentID := uuid.V7().String()
	p := path.Join(accessBasePath, enrollmentPath, enrollmentID)

	req := &createEnrollment{
		ID:           enrollmentID,
		OwnerID:      ownerID,
		ClientID:     s.client.clientID,
		ClientSecret: s.client.clientSecret,
	}
	e := new(Enrollment)
	return e, s.client.put(ctx, p, req, e)
}
