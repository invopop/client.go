package invopop

const (
	siloBasePath = "/silo/v1"
)

// SiloService implements the Invopop Silo API.
type SiloService service

// Entries provides a wrapper around silo entry methods.
func (svc *SiloService) Entries() *SiloEntriesService {
	return (*SiloEntriesService)(svc)
}

// Attachments provides a wrapper around silo entry attachment methods.
func (svc *SiloService) Attachments() *SiloAttachmentsService {
	return (*SiloAttachmentsService)(svc)
}

// Meta provides a wrapper around silo meta methods.
func (svc *SiloService) Meta() *SiloMetaService {
	return (*SiloMetaService)(svc)
}

// GOBL provides a wrapper around silo's gobl methods.
func (svc *SiloService) GOBL() *SiloGOBLService {
	return (*SiloGOBLService)(svc)
}

// Spool provides a wrapper for the Silo's Spool resource.
func (svc *SiloService) Spool() *SiloSpoolService {
	return (*SiloSpoolService)(svc)
}
