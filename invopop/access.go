package invopop

const (
	accessBasePath = "/access/v1"
)

// AccessService provides a wrapper around the Invopop Access public API.
type AccessService service

// Enrollment returns the service for Access Enrollments
func (svc *AccessService) Enrollment() *EnrollmentService {
	return (*EnrollmentService)(svc)
}

// Company returns the service for Access Companies
func (svc *AccessService) Company() *CompanyService {
	return (*CompanyService)(svc)
}
