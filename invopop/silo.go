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

// Meta provides a wrapper around silo meta methods.
func (svc *SiloService) Meta() *SiloMetaService {
	return (*SiloMetaService)(svc)
}

// GOBL provides a wrapper around silo's gobl methods.
func (svc *SiloService) GOBL() *SiloGOBLService {
	return (*SiloGOBLService)(svc)
}
