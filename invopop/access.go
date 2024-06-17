package invopop

const (
	accessBasePath = "/access/v1"
)

// AccessService provides a wrapper around the Invopop Access public API.
type AccessService struct {
	*service
	enrollment *EnrollmentService
}

func newAccessService(s *service) *AccessService {
	return &AccessService{
		service:    s,
		enrollment: (*EnrollmentService)(s),
	}
}

// Enrollment returns the service for Access Enrollments
func (svc *AccessService) Enrollment() *EnrollmentService {
	return svc.enrollment
}
