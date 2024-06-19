package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"path"
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

	Data json.RawMessage `json:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`

	Disabled bool `json:"disabled" title:"Disabled" description:"Whether the enrollment is disabled." example:"false"`

	Token string `json:"token" title:"Token" description:"A token that may be used to authenticate the enrollment with API operations."`
}

type authorizeEnrollment struct {
	OwnerID      string `json:"owner_id,omitempty" title:"Owner ID" description:"The ID of the entity that owns the enrollment. It is essential this is provided from a trusted source or an auth token is provided in the headers." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	ClientID     string `json:"client_id" title:"Client ID" description:"The ID of the application that is being enrolled." example:"01900e17-db4d-78a5-8505-c93ae63e8a0d"`
	ClientSecret string `json:"client_secret" title:"Client Secret" description:"The secret key of the application that is being enrolled." example:"01900e17-db4d-78a5-8505-c93ae63e8a0d"`
}

type updateEnrollment struct {
	Data json.RawMessage `param:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`
}

// Authorize tries to update the Enrollment object with an embedded token to use
// for subsequent requests to the API.
//
// OAuth credentials must have been configured in the client for this to work, and
// will be used alongside regular token authentication to ensure the client has
// the necessary permissions to generate the enrollment token.
func (s *EnrollmentService) Authorize(ctx context.Context, e *Enrollment) error {
	p := path.Join(accessBasePath, enrollmentPath, authorizePath)
	if s.client.clientID == "" || s.client.clientSecret == "" {
		return errors.New("missing OAuth client credentials: client_id and client_secret")
	}
	req := &authorizeEnrollment{
		OwnerID:      e.OwnerID, // may be empty if intention is to use token
		ClientID:     s.client.clientID,
		ClientSecret: s.client.clientSecret,
	}
	return s.client.post(ctx, p, req, e)
}

// Fetch retrieves an enrollment associated with the current authentication token.
func (s *EnrollmentService) Fetch(ctx context.Context, e *Enrollment) error {
	p := path.Join(accessBasePath, enrollmentPath)
	return s.client.get(ctx, p, e)
}

// Update allows applications to update their enrollment's embedded data.
func (s *EnrollmentService) Update(ctx context.Context, e *Enrollment) error {
	p := path.Join(accessBasePath, enrollmentPath)
	req := &updateEnrollment{
		Data: e.Data,
	}
	return s.client.post(ctx, p, req, e)
}
