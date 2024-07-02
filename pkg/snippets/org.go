package snippets

import (
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/uuid"
)

// OrgParty contains a snippet of the data for an org party.
type OrgParty struct {
	UUID    uuid.UUID        `json:"uuid,omitempty"`
	Name    string           `json:"name"`
	Alias   string           `json:"alias,omitempty"`
	Country l10n.CountryCode `json:"country,omitempty"`
	TaxCode string           `json:"tax_code,omitempty"`
}
