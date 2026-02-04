package invopop

const (
	accessBasePath = "/access/v1"
)

// AccessService provides a wrapper around the Invopop Access public API.
type AccessService service

// NewSession will instantiate a new session object and ensure that it references
// the client which is expected to have been prepared with the base OAuth App
// credentials.
func (s *AccessService) NewSession() *Session {
	sess := new(Session)
	sess.client = s.client
	return sess
}

// NewSessionWithToken will instantiate a new session object with the provided
// token value. This is a convenience method around NewSession and is intended to
// be used before making a call to Authorize to validate that the token is valid.
func (s *AccessService) NewSessionWithToken(token string) *Session {
	sess := s.NewSession()
	sess.SetToken(token)
	return sess
}

// NewSessionWithOwnerID will instantiate a new session object with the provided
// owner ID. This is a convenience method around NewSession and is intended to be
// used before making a call to Authorize to validate that the owner ID is valid.
func (s *AccessService) NewSessionWithOwnerID(ownerID string) *Session {
	sess := s.NewSession()
	sess.SetOwnerID(ownerID)
	return sess
}

// Enrollment returns the service for Access Enrollments
func (s *AccessService) Enrollment() *EnrollmentService {
	return (*EnrollmentService)(s)
}

// Workspace returns the service for Access Workspaces
func (s *AccessService) Workspace() *WorkspaceService {
	return (*WorkspaceService)(s)
}

// Company returns the service for Access Workspaces
// Deprecated: Use Workspace instead.
func (s *AccessService) Company() *WorkspaceService {
	return s.Workspace()
}

// Org returns the service for Access Organizations
func (s *AccessService) Org() *OrgService {
	return (*OrgService)(s)
}
