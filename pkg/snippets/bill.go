package snippets

import (
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/uuid"
)

// BillInvoice contains the summary details of an invoice.
type BillInvoice struct {
	// Key invoice header fields
	UUID            uuid.UUID     `json:"uuid,omitempty"`
	Type            string        `json:"type"`
	Series          string        `json:"series,omitempty"`
	Code            string        `json:"code"`
	Currency        currency.Code `json:"currency"`
	IssueDate       cal.Date      `json:"issue_date"`
	Supplier        *OrgParty     `json:"supplier,omitempty"`
	Customer        *OrgParty     `json:"customer,omitempty"`
	Total           *num.Amount   `json:"total,omitempty"`
	Tax             *num.Amount   `json:"tax,omitempty"`
	TotalWithTax    *num.Amount   `json:"total_with_tax,omitempty"`
	Payable         *num.Amount   `json:"payable,omitempty"`
	PrecedingSeries string        `json:"p_series,omitempty"`
	PrecedingCode   string        `json:"p_code,omitempty"`
	PrecedingUUID   uuid.UUID     `json:"p_uuid,omitempty"`
}
