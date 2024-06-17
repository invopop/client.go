package invopop

import (
	"context"
	"encoding/json"
	"path"
)

const (
	enrollmentPath = "/enrollment"
)

// EnrollmentService helps manage access to enrollments
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

// AuthorizeEnrollment defines the payload required to authorize an app using its client credentials
// and receive a token that can be used for subsequent requests.
type AuthorizeEnrollment struct {
	OwnerID      string `json:"owner_id,omitempty" title:"Owner ID" description:"The ID of the entity that owns the enrollment. It is essential this is provided from a trusted source or an auth token is provided in the headers." example:"347c5b04-cde2-11ed-afa1-0242ac120002"`
	ClientID     string `json:"client_id" title:"Client ID" description:"The ID of the application that is being enrolled." example:"01900e17-db4d-78a5-8505-c93ae63e8a0d"`
	ClientSecret string `json:"client_secret" title:"Client Secret" description:"The secret key of the application that is being enrolled." example:"01900e17-db4d-78a5-8505-c93ae63e8a0d"`
}

// UpdateEnrollment is used by applications to update the enrollment's embedded data.
type UpdateEnrollment struct {
	Data json.RawMessage `param:"data" title:"Data" description:"Additional data associated with the enrollment." example:"{\"key\":\"value\"}"`
}

// Authorize tries to provide an Enrollment object with an embedded token to use
// for subsequent requests to the API. This method will automatically update the
// the client's token if successful so that the same client can be re-used for all
// subsequent calls to the API.
func (s *EnrollmentService) Authorize(ctx context.Context, req *AuthorizeEnrollment) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath, "authorize")
	e := new(Enrollment)
	if err := s.client.post(ctx, p, req, e); err != nil {
		return nil, err
	}
	s.updateAuthToken(e)
	return e, nil
}

// Fetch retrieves an enrollment associated with the current authentication token.
func (s *EnrollmentService) Fetch(ctx context.Context) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath)
	e := new(Enrollment)
	if err := s.client.get(ctx, p, e); err != nil {
		return nil, err
	}
	s.updateAuthToken(e)
	return e, nil
}

// Update allows applications to update the enrollment's embedded data.
func (s *EnrollmentService) Update(ctx context.Context, req *UpdateEnrollment) (*Enrollment, error) {
	p := path.Join(accessBasePath, enrollmentPath)
	e := new(Enrollment)
	if err := s.client.post(ctx, p, req, e); err != nil {
		return nil, err
	}
	s.updateAuthToken(e)
	return e, nil
}

func (s *EnrollmentService) updateAuthToken(e *Enrollment) {
	if e.Token != "" {
		s.client.SetAuthToken(e.Token)
	}
}
